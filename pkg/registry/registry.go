package registry

import (
	"net/http"
	"sync"

	"github.com/emicklei/go-restful/v3"
	"gomod.alauda.cn/alauda-backend/pkg/server"
)

var (
	// DefaultRegistry default registry instance
	DefaultRegistry = NewRegistry()
)

// Add adds a webservice to the DefaultRegistry
func Add(ws *restful.WebService) {
	DefaultRegistry.Add(ws)
}

// AddBuilder adds a builder function to DefaultRegistry
func AddBuilder(f WebServiceBuilder) {
	DefaultRegistry.AddBuilder(f)
}

// AddHttpHandler adds a http.Handler function to DefaultRegistry
func AddHttpHandler(path string, handler http.Handler) {
	DefaultRegistry.AddHttpHandler(path, handler)
}

// WebServices return all webservices registered to the
// DefaultRegistry
func WebServices() []*restful.WebService {
	return DefaultRegistry.GetServices()
}

// HTTPHandlers return all httphandler of Defaultregistry
func HTTPHandlers() map[string]http.Handler {
	return DefaultRegistry.HTTPHandlers
}

// Builders return all registred builder functions of DefaultRegistry
func Builders() []WebServiceBuilder {
	return DefaultRegistry.GetBuilders()
}

// Build build all services
func Build(srv server.Server) (services []*restful.WebService, err error) {
	var ws *restful.WebService
	for _, builder := range Builders() {
		if ws, err = builder(srv); err != nil {
			return
		}
		services = append(services, ws)
	}
	return
}

// WebServiceBuilder builds and returns a webservice
type WebServiceBuilder func(svr server.Server) (*restful.WebService, error)

// Registry handles handler registration
type Registry struct {
	WebServices  []*restful.WebService
	Constructors []WebServiceBuilder
	HTTPHandlers map[string]http.Handler
	lock         sync.RWMutex
}

// NewRegistry inits a new registry for registration
func NewRegistry() *Registry {
	return &Registry{
		WebServices:  make([]*restful.WebService, 0, 10),
		Constructors: make([]WebServiceBuilder, 0, 10),
		HTTPHandlers: make(map[string]http.Handler),
		lock:         sync.RWMutex{},
	}
}

// AddBuilder adds a builder function to registry
func (r *Registry) AddBuilder(f WebServiceBuilder) {
	r.Constructors = append(r.Constructors, f)
}

// AddHttpHandler adds a http handler function to registry
func (r *Registry) AddHttpHandler(path string, handler http.Handler) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.HTTPHandlers[path] = handler
}

// Add adds a new webservice in the registry
func (r *Registry) Add(ws *restful.WebService) {
	r.WebServices = append(r.WebServices, ws)
}

// GetServices returns all registred services
func (r *Registry) GetServices() []*restful.WebService {
	return r.WebServices
}

// GetBuilders returns all registred constructors
func (r *Registry) GetBuilders() []WebServiceBuilder {
	return r.Constructors
}
