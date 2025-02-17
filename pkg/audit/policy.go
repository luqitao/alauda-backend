package audit

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v2"
	auditinternal "k8s.io/apiserver/pkg/apis/audit"
)

var defaultVerbMatching = map[string]string{
	"get":    "view",
	"put":    "update",
	"post":   "create",
	"delete": "delete",
}

// auditCtx is used to render ObjectReference's template
type auditCtx struct {
	Path     string
	Params   map[string][]string
	Request  interface{}
	Response interface{}
}

// CheckPolicyMatch checks if request matches audit policy, and return if matched and the matched rule
func CheckPolicyMatch(req *http.Request, policy *Policy) (matched bool, rule PolicyRule) {
	for _, r := range policy.Rules {
		reg, err := regexp.Compile(r.Match.Path)
		if err != nil {
			continue
		}
		if !reg.MatchString(req.URL.Path) {
			continue
		}
		requestMethod := strings.ToLower(req.Method)
		if !contains(r.Match.Methods, requestMethod) {
			continue
		}
		return true, r
	}
	return
}

// ExecutePolicyProcess will fullfill event's fieild accourding to specified rule
func ExecutePolicyProcess(e *Event, r PolicyRule, req *http.Request) {
	requestMethod := strings.ToLower(req.Method)
	verb, ok := r.Process.VerbMatching[requestMethod]
	if !ok {
		verb = defaultVerbMatching[requestMethod]
	}
	e.Verb = verb

	req.ParseForm()
	ctx := auditCtx{
		Path:     req.URL.Path,
		Params:   req.Form,
		Request:  e.RequestObject,
		Response: e.ResponseObject,
	}
	e.ObjectRef = &auditinternal.ObjectReference{
		Resource:    convertTemplate(r.Process.ObjectRef.Resource, ctx),
		Namespace:   convertTemplate(r.Process.ObjectRef.Namespace, ctx),
		Name:        convertTemplate(r.Process.ObjectRef.Name, ctx),
		APIGroup:    convertTemplate(r.Process.ObjectRef.APIGroup, ctx),
		APIVersion:  convertTemplate(r.Process.ObjectRef.APIVersion, ctx),
		Subresource: convertTemplate(r.Process.ObjectRef.SubResource, ctx),
	}

	switch r.Level {
	case auditinternal.LevelMetadata:
		e.Level = auditinternal.LevelMetadata
		emptyRequest := make(map[string]interface{})
		emptyResponse := make(map[string]interface{})
		e.RequestObject = &emptyRequest
		e.ResponseObject = &emptyResponse
	case auditinternal.LevelRequest:
		e.Level = auditinternal.LevelRequest
		emptyResponse := make(map[string]interface{})
		e.ResponseObject = &emptyResponse
	}

	return
}

// LoadPolicyFromFile generates a Policy object from a specified file
func LoadPolicyFromFile(filePath string) (*Policy, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path not specified")
	}
	policyDef, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file path %q: %+v", filePath, err)
	}

	ret, err := LoadPolicyFromBytes(policyDef)
	if err != nil {
		return nil, fmt.Errorf("%v: from file %v", err.Error(), filePath)
	}

	return ret, nil
}

// LoadPolicyFromBytes generates a Policy object from bytes
func LoadPolicyFromBytes(policyDef []byte) (*Policy, error) {
	policy := &Policy{}
	err := yaml.Unmarshal(policyDef, policy)
	if err != nil {
		return nil, fmt.Errorf("failed decoding: %v", err)
	}

	return policy, nil
}

// render a template string and return the rendered value
// if template string is invalid return the raw template string
// if template render failed return empty string
func convertTemplate(tmplStr string, ctx auditCtx) string {
	tmpl, err := template.New("objectRef").Funcs(sprig.FuncMap()).Parse(tmplStr)
	if err != nil {
		return tmplStr
	}
	raw := new(bytes.Buffer)
	err = tmpl.Execute(raw, ctx)
	if err != nil {
		return ""
	}
	return raw.String()
}

// check if string slice contains a specified item
func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}
