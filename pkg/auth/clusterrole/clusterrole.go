package clusterrole

import (
	"sync"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type Resolver struct {
	cache         map[string]rbacv1.ClusterRole
	lock          sync.RWMutex
	dynamicClient dynamic.Interface
}

func NewResolver(dynamicClient dynamic.Interface, stopCh <-chan struct{}) *Resolver {
	r := &Resolver{
		dynamicClient: dynamicClient,
		cache:         make(map[string]rbacv1.ClusterRole),
	}
	requirement, _ := labels.NewRequirement("auth.cpaas.io/role.relative", selection.Exists, nil)
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicClient, 10*time.Minute, metav1.NamespaceAll, func(options *metav1.ListOptions) {
		options.LabelSelector = labels.NewSelector().Add(*requirement).String()
	})
	informer := factory.ForResource(rbacv1.SchemeGroupVersion.WithResource("clusterroles"))
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if cluster, err := r.toClusterRole(obj); err == nil && cluster != nil {
				r.updateClusterRole(cluster)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			if cluster, err := r.toClusterRole(new); err == nil && cluster != nil {
				r.updateClusterRole(cluster)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if cluster, err := r.toClusterRole(obj); err == nil && cluster != nil {
				r.removeClusterRole(cluster)
			}
		},
	})

	factory.Start(stopCh)
	// Wait for informer to sync
	for _, r := range factory.WaitForCacheSync(stopCh) {
		if !r {
			panic("faild to start clusterrole informer")
		}
	}
	return r
}

func (r *Resolver) toClusterRole(obj interface{}) (*rbacv1.ClusterRole, error) {
	if obj == nil {
		return nil, nil
	}
	if unstruct, ok := obj.(*unstructured.Unstructured); ok {
		clusterRole := rbacv1.ClusterRole{}
		return &clusterRole, runtime.DefaultUnstructuredConverter.FromUnstructured(unstruct.Object, &clusterRole)
	}
	return nil, nil
}

func (r *Resolver) updateClusterRole(cr *rbacv1.ClusterRole) {
	if cr == nil {
		return
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.cache[cr.Name] = *cr
}

func (r *Resolver) removeClusterRole(cr *rbacv1.ClusterRole) {
	if cr == nil {
		return
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.cache, cr.Name)
}

func (r *Resolver) GetClusterRoles(roleRelativeName string) []rbacv1.ClusterRole {
	r.lock.RLock()
	defer r.lock.RUnlock()
	ret := []rbacv1.ClusterRole{}
	for _, v := range r.cache {
		relativeName, ok := v.Labels["auth.cpaas.io/role.relative"]
		if !ok {
			continue
		}
		if relativeName != roleRelativeName {
			continue
		}
		ret = append(ret, v)
	}
	return ret
}
