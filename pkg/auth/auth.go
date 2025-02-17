package auth

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	authv1 "gomod.alauda.cn/alauda-backend/pkg/auth/apis/v1"
	"gomod.alauda.cn/alauda-backend/pkg/auth/clusterrole"
	"gomod.alauda.cn/alauda-backend/pkg/auth/request"
	"gomod.alauda.cn/alauda-backend/pkg/auth/user"
	"gomod.alauda.cn/alauda-backend/pkg/auth/userbinding"
	"gomod.alauda.cn/alauda-backend/pkg/util/token"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const (
	ResProject      = "res:project"
	ResNamespace    = "res:ns"
	ResResourceName = "res:name"
	ResCluster      = "res:cluster"
)

type Permission struct {
	RoleName    string
	Actions     []string
	Constraints map[string]string
	Resource    schema.GroupResource
}

// IsInConstraint returns true if the permission constrains is equal or broader than the given
func (p *Permission) IsInConstraint(constraints map[string]string) bool {
	// no constrains, allows all
	if len(p.Constraints) == 0 {
		return true
	}
	if p.matchAllConstraints(constraints) {
		return true
	}

	return false
}

func (p *Permission) matchAllConstraints(constraints map[string]string) bool {
	//log.Printf("************p.Constraints: %+v, constraints: %+v", p.Constraints, constraints)

	for k, v := range p.Constraints {
		if val, ok := constraints[k]; !ok || v != val {
			//log.Printf(">>>>>>>constraints, k: %+v , val: %+v, v: %+v", k, val, v)

			clusters := []string{}
			if k == ResCluster {
				clusters = strings.Split(v, ",")
			}
			if StringInSlice(val, clusters) {
				return true
			}

			return false
		}
	}
	return true
}

type AuthManager struct {
	erebusService       string
	cache               Cache
	userBindingResolver *userbinding.Resolver
	userResolver        *user.Manager
	clusterRoleResolver *clusterrole.Resolver
	requestInfoResolver *request.RequestInfoFactory
}

func NewManager(cache Cache, erebusService string, userBinding *userbinding.Resolver, clusterRole *clusterrole.Resolver, requestInfoResolver *request.RequestInfoFactory, userResolver *user.Manager) Manager {
	mgr := &AuthManager{
		erebusService:       erebusService,
		cache:               cache,
		userBindingResolver: userBinding,
		clusterRoleResolver: clusterRole,
		requestInfoResolver: requestInfoResolver,
		userResolver:        userResolver,
	}
	return mgr
}

func (m *AuthManager) Authenticate(ctx context.Context, req *http.Request) error {
	// validate the format and existence of token
	_, err := token.ParseJWTFromHeader(req)
	if err != nil {
		return err
	}

	var cfg *rest.Config
	var client *rest.RESTClient
	cfg, err = rest.InClusterConfig()
	if err != nil {
		return err
	}
	rawToken, _ := token.ParseRawToken(req)
	fmt.Printf("m.erebusService: %s\n", m.erebusService)
	m.erebusService = "erebus.cpaas-system"
	cfg.Host = "https://" + strings.TrimRight(m.erebusService, "/") + "/kubernetes/global"
	// update in cluster config
	cfg.BearerToken = rawToken
	cfg.BearerTokenFile = ""
	cfg.TLSClientConfig = rest.TLSClientConfig{Insecure: true}

	if cfg.NegotiatedSerializer == nil {
		cfg.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	}

	client, err = rest.UnversionedRESTClientFor(cfg)
	if err != nil {
		return err
	}
	_, err = client.Get().AbsPath("/api").DoRaw(context.TODO())
	return nil
}

func (m *AuthManager) Authorize(ctx context.Context, req *http.Request, opt *FilterOption) (bool, error) {
	jwtToken, err := token.ParseJWTFromHeader(req)
	if err != nil {
		return false, err
	}
	fmt.Printf("req jwtToken: %+v", *jwtToken)

	if jwtToken.IsServiceAccount() {
		// TODO sa authz
		fmt.Printf("skip serviceaccount authz: %s", jwtToken.Subject)
		return true, nil
	}

	userEmailName := EmailToName(jwtToken.Email)
	requestInfo, err := m.requestInfoResolver.NewRequestInfo(req)
	fmt.Printf("user: %s requestInfo: %+v\n", userEmailName, *requestInfo)
	if err != nil {
		return false, err
	}

	// 资源映射
	if opt != nil && len(opt.ResourceMap) > 0 {
		gr := requestInfo.Resource + "." + requestInfo.APIGroup
		val, ok := opt.ResourceMap[gr]
		if ok {
			items := strings.Split(val, ".")
			requestInfo.Resource = items[0]
			requestInfo.APIGroup = strings.Join(items[1:], ".")
		}
	}

	resource := schema.GroupResource{
		Group:    requestInfo.APIGroup,
		Resource: requestInfo.Resource,
	}

	constraints := map[string]string{}
	if len(requestInfo.Project) > 0 {
		constraints[ResProject] = requestInfo.Project
	}
	if len(requestInfo.Namespace) > 0 {
		constraints[ResNamespace] = requestInfo.Namespace
	}
	if len(requestInfo.Cluster) > 0 {
		constraints[ResCluster] = requestInfo.Cluster
	}
	if len(requestInfo.Name) > 0 {
		constraints[ResResourceName] = requestInfo.Name
	}

	// fmt.Printf("---->>>resource: %+v, constraints: %+v \n", resource, constraints)

	verify, err := m.Verify(userEmailName, requestInfo.Verb, resource, constraints)
	if err != nil {
		return false, err
	}

	return verify, nil
}

// GetActionsForResourceFast.permission: {RoleName:namespace-admin-system Actions:[get list watch] Constraints:map[res:cluster:global res:ns:proj01 res:project:proj01] Resource:userbindings.auth.alauda.io}
// GetActionsForResourceFast.permission: {RoleName:namespace-admin-system Actions:[*] Constraints:map[res:cluster:global res:ns:proj01 res:project:proj01] Resource:userbindings.auth.alauda.io}
// rbac.Verify	{"user": "8bd108c8a01a892d129c52484ef97a0d", "resource": "userbindings.auth.alauda.io", "constraints": {"res:project":"proj01"}, "action": "create", "actions": []}
func (m *AuthManager) Verify(user string, action string, resource schema.GroupResource, constraints map[string]string) (bool, error) {
	actions, err := m.GetActions(user, resource, constraints, true)
	//logger.Info("rbac.Verify", zap.Any("user", user), zap.Any("resource", resource), zap.Any("constraints", constraints), zap.Any("action", action), zap.Any("actions", actions))
	if err != nil {
		return false, err
	}
	return hasAction(action, actions), nil
}

func (m *AuthManager) GetActions(user string, resource schema.GroupResource, constraints map[string]string, useCache bool) ([]string, error) {
	if constraints == nil {
		constraints = map[string]string{}
	}

	if user == "" {
		return nil, errors.NewBadRequest("user is empty")
	}

	if resource.Resource == "" {
		return nil, errors.NewBadRequest("resource is empty")
	}

	// use cache
	if useCache {

	}

	userPerms, err := m.GetUserPermissions(user, resource)
	if err != nil {
		return nil, err
	}

	actions := m.GetActionsForResourceFast(userPerms, constraints)

	// fmt.Printf("-------> actions: %+v\n", actions)
	return actions, nil
}

func (m *AuthManager) GetUserPermissions(user string, resource schema.GroupResource) ([]*Permission, error) {
	userbindings := m.userBindingResolver.GetUserBindings(user)
	if u, err := m.userResolver.Get(user); err == nil {
		if u.Spec.Groups != nil && len(u.Spec.Groups) > 0 {
			groupUserbindings := m.userBindingResolver.GetUserBindingsByGroups(u.Spec.Groups)
			if len(groupUserbindings) > 0 {
				userbindings = append(userbindings, groupUserbindings...)
			}
		}
	}

	permissions := make([]*Permission, 0)
	if userbindings == nil {
		return permissions, nil
	}

	for _, binding := range userbindings {
		ch := make(chan []*Permission)
		go func(chnl chan []*Permission) {
			perms, err := m.GetUserBindingPermissions(&binding, resource)
			if err != nil {
				//logger.Error("GetUserBindingPermissions", zap.Any("resource", resource), zap.Any("userbinding", binding))
			}
			chnl <- perms
			close(chnl)
		}(ch)
		for perms := range ch {
			permissions = append(permissions, perms...)
		}
	}
	// fmt.Printf("-------> permissions: %+v\n", permissions)

	return permissions, nil
}

func (m *AuthManager) GetUserBindingPermissions(userbinding *authv1.UserBinding, resource schema.GroupResource) ([]*Permission, error) {
	//k8sClient, err := apiHandler.cManager.Client(request)
	//if err != nil {
	//	return nil, err
	//}
	//labelSelector := fmt.Sprintf("%s,%s=%s", constant.LabelRoleRelative, constant.LabelRoleRelative, userbinding.RoleName())
	//opts := metav1.ListOptions{
	//	LabelSelector: labelSelector,
	//}
	//clusterRoleList, err := k8sClient.RbacV1().ClusterRoles().List(context.TODO(), opts)
	//if err != nil {
	//	return nil, err
	//}
	clusterRoles := m.clusterRoleResolver.GetClusterRoles(userbinding.RoleName())
	permissions := make([]*Permission, 0)
	for _, clusterRole := range clusterRoles {
		ch := make(chan *Permission)
		go func(clusterRole *rbacv1.ClusterRole, userbinding *authv1.UserBinding, resource schema.GroupResource, chnl chan *Permission) {
			for _, rule := range clusterRole.Rules {
				// loop apigroup
				for _, group := range rule.APIGroups {

					if resource.Group != group && group != "*" {
						continue
					}

					// loop resources
					for _, res := range rule.Resources {

						if resource.Resource != res && res != "*" {
							continue
						}

						// loop resourceNames
						if len(rule.ResourceNames) > 0 {
							for _, resourceName := range rule.ResourceNames {

								chnl <- NewPermission(userbinding, resource, rule.Verbs, resourceName)
							}
						} else {
							chnl <- NewPermission(userbinding, resource, rule.Verbs, "")
						}

					} // end resources

				} // end apigroup
			}
			close(chnl)

		}(&clusterRole, userbinding, resource, ch)

		for perm := range ch {
			permissions = append(permissions, perm)
		}
	}
	return permissions, nil
}

func NewPermission(userbinding *authv1.UserBinding, resource schema.GroupResource, actions []string, resourceName string) *Permission {
	constraints := make(map[string]string)
	if len(userbinding.ProjectName()) > 0 {
		constraints[ResProject] = userbinding.ProjectName()
	}
	if len(userbinding.NamespaceName()) > 0 {
		constraints[ResNamespace] = userbinding.NamespaceName()
	}
	if len(resourceName) > 0 {
		constraints[ResResourceName] = resourceName
	}
	if len(userbinding.NamespaceCluster()) > 0 {
		constraints[ResCluster] = userbinding.NamespaceCluster()
	}
	if userbinding.Spec.Scope == authv1.UserBindingScopeCluster {
		clusters := []string{}
		for _, c := range userbinding.Spec.Constraint {
			clusters = append(clusters, c.Cluster)
		}
		constraints[ResCluster] = strings.Join(clusters, ",")
	}

	return &Permission{
		RoleName:    userbinding.RoleName(),
		Actions:     actions,
		Constraints: constraints,
		Resource:    resource,
	}
}

func (m *AuthManager) GetActionsForResourceFast(permissions []*Permission, constraints map[string]string) []string {
	var (
		actionsMap = make(map[string]struct{})
		place      = struct{}{}
		actions    []string
	)

	for _, p := range permissions {
		//log.Printf("GetActionsForResourceFast.permission: %+v", *p)
		if p.IsInConstraint(constraints) {
			for _, a := range p.Actions {
				actionsMap[a] = place
			}
		}
	}

	actions = make([]string, 0)
	for k := range actionsMap {
		if k == "*" {
			actions = []string{"*"}
			break
		}
		actions = append(actions, k)
	}
	return actions
}

// hasAction - verifies if the string is in the slice
// action: action string e.g service:create
// allowed: slice of allowed actions e.g. service:create, service:update, etc..
func hasAction(action string, allowed []string) bool {
	for _, a := range allowed {
		if a == "*" {
			return true
		}
		if a == action {
			return true
		}
	}
	return false
}

func EmailToName(email string) string {
	if len(email) == 0 {
		return ""
	}
	name := email
	name = GetMD5Hash(name)
	name = strings.ToLower(name)
	return name
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
