package userbinding

import (
	"sync"
	"time"

	authv1 "gomod.alauda.cn/alauda-backend/pkg/auth/apis/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type Resolver struct {
	cache  map[string]map[string]authv1.UserBinding
	gcache map[string]map[string]authv1.UserBinding

	lock          sync.RWMutex
	dynamicClient dynamic.Interface
}

func NewResolver(dynamicClient dynamic.Interface, stopCh <-chan struct{}) *Resolver {
	r := &Resolver{
		dynamicClient: dynamicClient,
		cache:         make(map[string]map[string]authv1.UserBinding),
		gcache:        make(map[string]map[string]authv1.UserBinding),
	}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicClient, 10*time.Minute, metav1.NamespaceAll, nil)
	informer := factory.ForResource(authv1.SchemeGroupVersion.WithResource("userbindings"))
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if binding, err := r.toUserBinding(obj); err == nil && binding != nil {
				r.updateUserBinding(binding)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			if binding, err := r.toUserBinding(new); err == nil && binding != nil {
				r.updateUserBinding(binding)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if binding, err := r.toUserBinding(obj); err == nil && binding != nil {
				r.removeUserBinding(binding)
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

func (r *Resolver) toUserBinding(obj interface{}) (*authv1.UserBinding, error) {
	if obj == nil {
		return nil, nil
	}
	if unstruct, ok := obj.(*unstructured.Unstructured); ok {
		userbinding := authv1.UserBinding{}
		return &userbinding, runtime.DefaultUnstructuredConverter.FromUnstructured(unstruct.Object, &userbinding)
	}
	return nil, nil
}

func (r *Resolver) updateUserBinding(cr *authv1.UserBinding) {
	if cr == nil {
		return
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	if len(cr.UserEmailName()) > 0 {
		data, ok := r.cache[cr.UserEmailName()]
		if !ok {
			data = make(map[string]authv1.UserBinding)
		}
		data[cr.Name] = *cr
		r.cache[cr.UserEmailName()] = data
	}

	if len(cr.GroupName()) > 0 {
		data, ok := r.gcache[cr.GroupName()]
		if !ok {
			data = make(map[string]authv1.UserBinding)
		}
		data[cr.Name] = *cr
		r.gcache[cr.GroupName()] = data
	}
}

func (r *Resolver) removeUserBinding(cr *authv1.UserBinding) {
	if cr == nil {
		return
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	if len(cr.UserEmailName()) > 0 {
		if data, ok := r.cache[cr.UserEmailName()]; ok {
			_, ok = data[cr.Name]
			if ok {
				delete(data, cr.Name)
			}
			if len(data) == 0 {
				delete(r.cache, cr.UserEmailName())
			} else {
				r.cache[cr.UserEmailName()] = data
			}
		}
	}

	if len(cr.GroupName()) > 0 {
		if data, ok := r.gcache[cr.GroupName()]; ok {
			_, ok = data[cr.Name]
			if ok {
				delete(data, cr.Name)
			}
			if len(data) == 0 {
				delete(r.gcache, cr.GroupName())
			} else {
				r.gcache[cr.GroupName()] = data
			}
		}
	}
}

func (r *Resolver) GetUserBindings(user string) []authv1.UserBinding {
	r.lock.RLock()
	defer r.lock.RUnlock()
	ret := []authv1.UserBinding{}
	data, ok := r.cache[user]
	if !ok {
		return ret
	}
	for _, v := range data {
		ret = append(ret, v)
	}
	return ret
}

func (r *Resolver) GetUserBindingsByGroups(groups []string) []authv1.UserBinding {
	r.lock.RLock()
	defer r.lock.RUnlock()
	ret := []authv1.UserBinding{}

	for _, group := range groups {
		data, ok := r.gcache[group]
		if !ok {
			continue
		}
		for _, v := range data {
			ret = append(ret, v)
		}
	}
	return ret
}
