package decorator

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"gomod.alauda.cn/alauda-backend/pkg/auth"
	"gomod.alauda.cn/alauda-backend/pkg/server"
	"k8s.io/apimachinery/pkg/api/errors"
)

type Auth struct {
	server.Server
}

func NewAuth(srv server.Server) Auth {
	return Auth{Server: srv}
}

func (a Auth) AuthenticationFilter(req *restful.Request, res *restful.Response, chain *restful.FilterChain) {
	err := a.GetAuthManager().Authenticate(req.Request.Context(), req.Request)
	if err != nil {
		switch t := err.(type) {
		case errors.APIStatus:
			code := int(t.Status().Code)
			if code == 0 {
				code = http.StatusInternalServerError
			}
			res.WriteHeader(code)
			res.WriteAsJson(t)
		default:
			res.WriteError(http.StatusInternalServerError, err)
		}
		return
	}
	chain.ProcessFilter(req, res)
}

func (a Auth) AuthorizationFilter(opts ...auth.FilterOption) restful.FilterFunction {
	return func(req *restful.Request, res *restful.Response, chain *restful.FilterChain) {
		var opt *auth.FilterOption
		if len(opts) > 0 {
			opt = &opts[0]
		}
		verify, err := a.GetAuthManager().Authorize(req.Request.Context(), req.Request, opt)
		if err != nil {
			res.WriteError(http.StatusInternalServerError, err)
			return
		}
		if !verify {
			res.WriteError(http.StatusForbidden, fmt.Errorf("no permissions."))
			return
		}
		chain.ProcessFilter(req, res)
	}
}

func (a Auth) AuthFilter(opts ...auth.FilterOption) restful.FilterFunction {
	return func(req *restful.Request, res *restful.Response, chain *restful.FilterChain) {
		err := a.GetAuthManager().Authenticate(req.Request.Context(), req.Request)
		if err != nil {
			res.WriteError(http.StatusUnauthorized, err)
			return
		}
		var opt *auth.FilterOption
		if len(opts) > 0 {
			opt = &opts[0]
		}
		verify, err := a.GetAuthManager().Authorize(req.Request.Context(), req.Request, opt)
		if err != nil {
			res.WriteError(http.StatusInternalServerError, err)
			return
		}
		if !verify {
			res.WriteError(http.StatusForbidden, fmt.Errorf("no permissions."))
			return
		}
		chain.ProcessFilter(req, res)
	}
}
