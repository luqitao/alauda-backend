package audit

import (
	"net/http"
	"strings"
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
func GetToken(req *http.Request) (token string) {
	authHeader := req.Header.Get(AuthorizationHeader)
	if authHeader != "" && strings.HasPrefix(authHeader, BearerPrefix) && strings.TrimPrefix(authHeader, BearerPrefix) != "" {
		token = strings.TrimPrefix(authHeader, BearerPrefix)
		return
	}
	return
}
