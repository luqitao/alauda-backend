package decorator

import (
	"net/http"

	restful "github.com/emicklei/go-restful/v3"
	"gomod.alauda.cn/alauda-backend/pkg/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	generator = NewWSGenerator()
)

// NewWebService constructs a new webservice
func NewWebService(srv server.Server) *restful.WebService {
	ws := generator.New(srv)
	return ws
}

// WithBadRequest bad request decorator
func WithBadRequest(builder *restful.RouteBuilder) *restful.RouteBuilder {
	return builder.Returns(http.StatusBadRequest, "BadRequest", metav1.Status{})
}

// WithUnauthorized Unauthorized decorator
func WithUnauthorized(builder *restful.RouteBuilder) *restful.RouteBuilder {
	return builder.Returns(http.StatusUnauthorized, "Unauthorized", metav1.Status{})
}

// WithInternalServerError Unauthorized decorator
func WithInternalServerError(builder *restful.RouteBuilder) *restful.RouteBuilder {
	return builder.Returns(http.StatusInternalServerError, "InternalServerError", metav1.Status{})
}

// WithAuth adds InternalServerError and Unauthorized
func WithAuth(builder *restful.RouteBuilder) *restful.RouteBuilder {
	return WithUnauthorized(
		WithInternalServerError(builder),
	)
}

// WithAuthAndBadRequest adds WithAuth and WithBadRequest
func WithAuthAndBadRequest(builder *restful.RouteBuilder) *restful.RouteBuilder {
	return WithBadRequest(WithAuth(builder))
}

// WebServiceGenerator generator interface for webservices
type WebServiceGenerator interface {
	New(server.Server) *restful.WebService
}

// NewWSGenerator constructs a webservice generator
func NewWSGenerator() *DefaultWebServiceGen {
	return &DefaultWebServiceGen{}
}

// DefaultWebServiceGen default generator for webservices
type DefaultWebServiceGen struct{}

// New builds a new default webservice, useful to add some global parameters
// like content-type and authorization
func (g *DefaultWebServiceGen) New(srv server.Server) *restful.WebService {
	ws := new(restful.WebService)
	ws.Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Param(restful.HeaderParameter("Authorization", "Given Bearer token will use this as authorization for the API"))

	return ws
}
