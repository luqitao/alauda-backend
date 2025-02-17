package client

import (
	"sync"
	"time"

	dc "github.com/alauda/cyborg/pkg/client"
	restful "github.com/emicklei/go-restful/v3"
	"gomod.alauda.cn/alauda-backend/pkg/util/hash"
	"gomod.alauda.cn/log"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	expirationTime = time.Minute * 60
	//DefaultInterval interval of 60 * minutes
	DefaultInterval = 60 * time.Minute
)

// DefaultManager default client manager
type DefaultManager struct {
	// Config configuration for manager
	config *Config

	// Configuration generation for secure authorized requests
	ConfigGeneratorFuncs []ConfigGenFunc

	// InsecureConfigGeneratorFuncs list of configuration generation methods
	// for insecure clients
	InsecureConfigGeneratorFuncs []ConfigGenFunc

	// clients map of k8s client
	clients    map[string]ClientEntity
	clientLock *sync.RWMutex

	onceWatch sync.Once
}

// ClientEntity is the interface for client to refresh cache expired time
type ClientEntity interface {
	Refresh()
	IsExpired() bool
}

type clientEntity struct {
	client       *kubernetes.Interface
	latestCalled time.Time
}

func (client *clientEntity) Refresh() {
	client.latestCalled = time.Now()
}

func (client *clientEntity) IsExpired() bool {
	return time.Since(client.latestCalled) > expirationTime
}

type dynamicClientEntity struct {
	dynamicClient *dynamic.NamespaceableResourceInterface

	latestCalled time.Time
}

func (client *dynamicClientEntity) Refresh() {
	client.latestCalled = time.Now()
}

func (client *dynamicClientEntity) IsExpired() bool {
	return time.Since(client.latestCalled) > expirationTime
}

var _ Manager = &DefaultManager{}

// NewManager inits a manager
func NewManager() *DefaultManager {
	return &DefaultManager{
		// GeneratorFuncs:               []GeneratorFunc{},
		ConfigGeneratorFuncs: []ConfigGenFunc{},
		// InsecureGeneratorFuncs:       []GeneratorFunc{},
		InsecureConfigGeneratorFuncs: []ConfigGenFunc{},

		clients:    make(map[string]ClientEntity),
		clientLock: &sync.RWMutex{},
	}
}

// WithConfig sets a configuration to manager
func (m *DefaultManager) WithConfig(config *Config) *DefaultManager {
	m.config = config
	return m
}

// With adds client configuration generator
func (m *DefaultManager) With(gen ...ConfigGenFunc) *DefaultManager {
	m.ConfigGeneratorFuncs = append(m.ConfigGeneratorFuncs, gen...)
	return m
}

// WithInsecure adds insecure client configuration generator
func (m *DefaultManager) WithInsecure(gen ...ConfigGenFunc) *DefaultManager {
	m.InsecureConfigGeneratorFuncs = append(m.InsecureConfigGeneratorFuncs, gen...)
	return m
}

// InsecureClient returns InCluster configuration client
// using pod's service account or according to kubeconfig during init
func (m *DefaultManager) InsecureClient() (client kubernetes.Interface, err error) {
	if m.config == nil || len(m.InsecureConfigGeneratorFuncs) == 0 {
		err = errors.NewUnauthorized("No client configuration provided")
		return
	}
	config, err := m.genConfig(nil, m.InsecureConfigGeneratorFuncs...)
	if err != nil {
		m.config.Log.Error("insecure client generation config failed", log.Err(err))
		return
	}
	client, err = m.genClient(config)
	if err != nil {
		m.config.Log.Error("insecure client generation failed", log.Err(err))
	}
	return
}

// Client returns a client given a request authorization options
func (m *DefaultManager) Client(req *restful.Request) (client kubernetes.Interface, err error) {
	if m.config == nil || len(m.ConfigGeneratorFuncs) == 0 {
		err = errors.NewUnauthorized("No client configuration provided")
		return
	}

	config, err := m.genConfig(req)
	if err != nil {
		m.config.Log.Error("secure client generation config failed", log.Err(err))
		return
	}
	client, err = m.GetClient(config)
	if err != nil {
		m.config.Log.Error("secure client generation failed", log.Err(err))
	}
	return
}

// Config gives an authenticated rest config given request
func (m *DefaultManager) Config(req *restful.Request) (config *rest.Config, err error) {
	return m.genConfig(req)
}

func (m *DefaultManager) genConfig(req *restful.Request, genFuncs ...ConfigGenFunc) (config *rest.Config, err error) {
	if len(genFuncs) == 0 {
		genFuncs = m.ConfigGeneratorFuncs
	}
	for _, gen := range genFuncs {
		config, err = gen(m.config, req)
		if err == nil && config != nil {
			m.config.setupConfig(config, err)
			return
		}
	}

	if config == nil {
		// no genconfigfunc successed
		err = errors.NewUnauthorized("config generation failed")

		m.config.Log.Error("no genFuncs triggered", log.Err(err))
	}
	return
}

// DynamicClient genreates a dynamic client instance
func (m *DefaultManager) DynamicClient(req *restful.Request, gvk *schema.GroupVersionKind) (client dynamic.NamespaceableResourceInterface, err error) {
	if m.config == nil || len(m.ConfigGeneratorFuncs) == 0 {
		err = errors.NewUnauthorized("No client configuration provided")
		return
	}

	config, err := m.genConfig(req)
	if err != nil {
		m.config.Log.Error("dynamic client generation config failed", log.Err(err))
		return
	}
	client, err = m.GetDynamicClient(config, gvk)
	if err != nil {
		m.config.Log.Error("dynamic client generation failed", log.Err(err))
	}
	return
}

func (m *DefaultManager) genClient(config *rest.Config) (client kubernetes.Interface, err error) {
	client, err = kubernetes.NewForConfig(config)
	return
}

func (m *DefaultManager) genDynamicClient(gvk *schema.GroupVersionKind, config *rest.Config) (client dynamic.NamespaceableResourceInterface, err error) {
	var (
		cyborgClient *dc.KubeClient
	)

	gv := gvk.GroupVersion()
	config.GroupVersion = &gv

	if gv.String() == "v1" {
		config.APIPath = "/api"
	} else {
		config.APIPath = "/apis"
	}
	cyborgClient, err = dc.NewKubeClient(config, "default")
	if err != nil || cyborgClient == nil {
		return
	}
	client, err = cyborgClient.ClientForGVK(*gvk)
	return
}

// ManagerConfig returns a clone of manager's configuration.
func (m *DefaultManager) ManagerConfig() Config {
	return *m.config
}

// Hash return a hash key for client base on config and gvk
func (m *DefaultManager) Hash(config *rest.Config, gvk *schema.GroupVersionKind) (hashcode string) {
	hashcode = hash.HashToString(config)

	if gvk != nil {
		hashGvk := hash.HashToString(gvk)

		hashcode = hashcode + hashGvk
	}
	return
}

// GetClient return client from cache or gen new client and put it to cache
func (m *DefaultManager) GetClient(config *rest.Config) (client kubernetes.Interface, err error) {
	// first call start watch
	m.watch()

	hashstr := m.Hash(config, nil)

	m.clientLock.RLock()
	cl, ok := m.clients[hashstr]
	m.clientLock.RUnlock()

	if ok {
		// cache hit refresh latesttime
		cl.Refresh()
		clt := cl.(*clientEntity)
		return *clt.client, nil
	}

	client, err = m.genClient(config)
	if err != nil {
		return
	}

	m.clientLock.Lock()
	m.clients[hashstr] = &clientEntity{client: &client, latestCalled: time.Now()}
	m.clientLock.Unlock()

	return
}

// GetDynamicClient return dynamic client from cache or gen new dynamic client and put it to cache.
func (m *DefaultManager) GetDynamicClient(config *rest.Config, gvk *schema.GroupVersionKind) (client dynamic.NamespaceableResourceInterface, err error) {
	// first call start watch
	m.watch()

	hashstr := m.Hash(config, gvk)

	m.clientLock.RLock()
	cl, ok := m.clients[hashstr]
	m.clientLock.RUnlock()
	if ok {
		// cache hit refresh latesttime
		cl.Refresh()
		clt := cl.(*dynamicClientEntity)
		return *clt.dynamicClient, nil
	}

	client, err = m.genDynamicClient(gvk, config)

	if err != nil {
		return
	}

	m.clientLock.Lock()
	m.clients[hashstr] = &dynamicClientEntity{dynamicClient: &client, latestCalled: time.Now()}
	m.clientLock.Unlock()

	return
}

func (m *DefaultManager) watch() {
	m.onceWatch.Do(func() {
		go m.watchClients(DefaultInterval)
	})
}

func (m *DefaultManager) watchClients(interval time.Duration) {
	ticker := time.NewTicker(interval)
	deletedHashs := []string{}
	for {
		select {
		case <-ticker.C:
			m.clientLock.RLock()
			for hashcode, client := range m.clients {
				if client.IsExpired() {
					deletedHashs = append(deletedHashs, hashcode)
					//delete(m.clients, hashcode)
				}
			}
			m.clientLock.RUnlock()

			// delete clients
			m.clientLock.Lock()
			for _, hashcode := range deletedHashs {
				delete(m.clients, hashcode)
			}
			m.clientLock.Unlock()
		}
	}
}

// NewRestClientForAPI gen restclient
func (m *DefaultManager) NewRestClientForAPI(fromCfg *rest.Config, gvk schema.GroupVersionKind, scheme *runtime.Scheme) (*rest.RESTClient, error) {
	groupVersion := gvk.GroupVersion()
	cfg := rest.Config{
		Host:    fromCfg.Host,
		APIPath: dynamic.LegacyAPIPathResolverFunc(gvk),
		ContentConfig: rest.ContentConfig{
			GroupVersion:         &groupVersion,
			NegotiatedSerializer: serializer.NewCodecFactory(scheme).WithoutConversion(),
			ContentType:          runtime.ContentTypeJSON,
		},
		BearerToken:     fromCfg.BearerToken,
		TLSClientConfig: fromCfg.TLSClientConfig,
		QPS:             fromCfg.QPS,
		Burst:           fromCfg.Burst,
	}
	return rest.RESTClientFor(&cfg)
}
