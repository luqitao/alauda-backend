package client

import (
	"strings"

	restful "github.com/emicklei/go-restful/v3"
)

const (
	// AuthorizationHeader authorization header for http requests
	AuthorizationHeader = "Authorization"
	// BearerPrefix bearer token prefix for token
	BearerPrefix = "Bearer "

	// QueryParameterTokenName authorization token for http requests
	QueryParameterTokenName = "token"
)

// GetToken get token from request headers or request query parameters.
// return emtry if no token find
func GetToken(req *restful.Request) (token string) {
	authHeader := req.HeaderParameter(AuthorizationHeader)

	if authHeader != "" && strings.HasPrefix(authHeader, BearerPrefix) && strings.TrimPrefix(authHeader, BearerPrefix) != "" {
		token = strings.TrimPrefix(authHeader, BearerPrefix)
		return
	}

	token = req.QueryParameter(QueryParameterTokenName)
	return
}

// GetClusterName get cluster name from request path or request query parameters.
func GetClusterName(multiClusterParameterName string, req *restful.Request) (clusterName string) {
	clusterName = req.PathParameter(multiClusterParameterName)
	if clusterName == "" {
		clusterName = req.QueryParameter(multiClusterParameterName)
	}

	return
}
