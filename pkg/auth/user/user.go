package user

import (
	"sync"
	"time"

	authv1 "gomod.alauda.cn/alauda-backend/pkg/auth/apis/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	kcache "k8s.io/client-go/tools/cache"
)

type Manager struct {
	cache map[string]*authv1.User
	lock  sync.RWMutex
}

func NewResolver(stopCh <-chan struct{}) *Manager {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	m := &Manager{
		cache: make(map[string]*authv1.User),
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clientset, 2*time.Hour, metav1.NamespaceAll, nil)

	userInformer := factory.ForResource(authv1.SchemeGroupVersion.WithResource("users")).Informer()

	go factory.Start(stopCh)
	for _, r := range factory.WaitForCacheSync(stopCh) {
		if !r {
			panic("faild to start users informer")
		}
	}

	userInformer.AddEventHandler(kcache.ResourceEventHandlerFuncs{
		AddFunc:    m.onAdd,
		UpdateFunc: m.onUpdate,
		DeleteFunc: m.onDelete,
	})
	return m
}

func (m *Manager) onAdd(obj interface{}) {
	if u, err := m.toUser(obj); err == nil && u != nil {
		m.addUser(u)
	}
}

func (m *Manager) onUpdate(oldObj, newObj interface{}) {
	if u, err := m.toUser(newObj); err == nil && u != nil {
		m.updateUser(u)
	}
}

func (m *Manager) onDelete(obj interface{}) {
	if u, err := m.toUser(obj); err == nil && u != nil {
		m.removeUser(u)
	}
}

func (m *Manager) Get(name string) (*authv1.User, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	u, ok := m.cache[name]
	if !ok {
		gv := schema.GroupResource{Group: authv1.SchemeGroupVersion.Group, Resource: "User"}
		return nil, errors.NewNotFound(gv, name)
	}
	return u, nil
}

func (m *Manager) toUser(obj interface{}) (*authv1.User, error) {
	if obj == nil {
		return nil, nil
	}
	if unstruct, ok := obj.(*unstructured.Unstructured); ok {
		u := authv1.User{}
		return &u, runtime.DefaultUnstructuredConverter.FromUnstructured(unstruct.Object, &u)
	}
	return nil, nil
}

func (m *Manager) removeUser(u *authv1.User) {
	if u == nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.cache, u.Name)
}

func (m *Manager) addUser(u *authv1.User) {
	if u == nil {
		return
	}
	if u.Spec.Groups == nil || len(u.Spec.Groups) == 0 {
		return
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	m.cache[u.Name] = u
}

func (m *Manager) updateUser(u *authv1.User) {
	if u == nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	if u.Spec.Groups == nil || len(u.Spec.Groups) == 0 {
		delete(m.cache, u.Name)
		return
	}
	m.cache[u.Name] = u
}
