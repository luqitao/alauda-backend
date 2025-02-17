package decorator

import (
	restful "github.com/emicklei/go-restful/v3"
	"gomod.alauda.cn/alauda-backend/pkg/context"
	"gomod.alauda.cn/alauda-backend/pkg/server"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Client client decorator. Used to initiates a kubernetes.Interface client
// and sets into the context. InsecureFilter can be used to set an ServiceAccount client
// SecureFilter can be used to use user credentials on the client
type Client struct {
	server.Server
}

// NewClient constructor for the Client decorator
func NewClient(srv server.Server) Client {
	return Client{Server: srv}
}

// InsecureFilter filter to populate the context with a ServiceAccount client
// will handle errors and may return errors on request
func (d Client) InsecureFilter(req *restful.Request, res *restful.Response, chain *restful.FilterChain) {
	client, err := d.GetManager().InsecureClient()
	d.SetClientContext(client, err, req, res, chain)
}

// SecureFilter filter to populate the context with a user's Bearer token client
// will handle errors and may return errors on request
func (d Client) SecureFilter(req *restful.Request, res *restful.Response, chain *restful.FilterChain) {
	client, err := d.GetManager().Client(req)
	d.SetClientContext(client, err, req, res, chain)
}

// DynamicFilterGenerator returns a filter for generating DynamicClient based on
// a given GroupVersionKind parameter
func (d Client) DynamicFilterGenerator(gvk *schema.GroupVersionKind) restful.FilterFunction {
	return func(req *restful.Request, res *restful.Response, chain *restful.FilterChain) {
		client, err := d.GetManager().DynamicClient(req, gvk)
		d.SetDynamicClientContext(client, err, req, res, chain)
	}
}

// SetClientContext given a client and an error will create the common logic to handle error
// and populate the context with the client
func (d Client) SetClientContext(client kubernetes.Interface, err error, req *restful.Request, res *restful.Response, chain *restful.FilterChain) {
	if err != nil || client == nil {
		d.Server.HandleError(err, req, res)
		return
	}
	req.Request = req.Request.WithContext(context.WithClient(req.Request.Context(), client))
	chain.ProcessFilter(req, res)
}

// SetDynamicClientContext given a client and an error will create the common logic to handle error
// and populate the context with the client
func (d Client) SetDynamicClientContext(client dynamic.NamespaceableResourceInterface, err error, req *restful.Request, res *restful.Response, chain *restful.FilterChain) {
	if err != nil || client == nil {
		d.Server.HandleError(err, req, res)
		return
	}
	req.Request = req.Request.WithContext(context.WithDynamicClient(req.Request.Context(), client))
	chain.ProcessFilter(req, res)
}
