package client

import (
	restful "github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Manager a client generator for k8s.io/client-go
type Manager interface {
	// InsecureClient returns InCluster configuration client
	// using pod's service account or according to kubeconfig during init
	InsecureClient() (kubernetes.Interface, error)

	// Client returns a client given a request authorization options
	Client(req *restful.Request) (kubernetes.Interface, error)

	// Generates a secure config according to the request
	Config(req *restful.Request) (*rest.Config, error)

	// DynamicClient generates a dynamic client based on request and GroupVersionKind information
	DynamicClient(req *restful.Request, gvk *schema.GroupVersionKind) (client dynamic.NamespaceableResourceInterface, err error)

	// ManagerConfig returns a clone of manager's configuration.
	ManagerConfig() Config

	// GetClient returns a client base on config
	GetClient(config *rest.Config) (client kubernetes.Interface, err error)

	// DynamicClient returns a dynamic client based on config and GroupVersionKind information
	GetDynamicClient(config *rest.Config, gvk *schema.GroupVersionKind) (client dynamic.NamespaceableResourceInterface, err error)
}

// GeneratorFunc generates a client given a configuration and a request
type GeneratorFunc func(cfg *Config, req *restful.Request) (kubernetes.Interface, error)

// ConfigGenFunc func(cfg creq, restful BadExpr) (*rest.Config, error)
// ConfigGenFunc generates a configuration for a given request
type ConfigGenFunc func(cfg *Config, req *restful.Request) (*rest.Config, error)
