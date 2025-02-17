package auth

import (
	"context"
	"net/http"
	"time"
)

type Manager interface {
	Authenticate(ctx context.Context, req *http.Request) error
	Authorize(ctx context.Context, req *http.Request, opt *FilterOption) (bool, error)
}

type Cache interface {
	GetAuthorize(key string) (bool, error)
	SetAuthorize(key string, val interface{}, d time.Duration) error
}

type FilterOption struct {
	// eg. abc.alauda.io:metrics.alauda.io, xyz.alauda.io:metrics.alauda.io
	ResourceMap map[string]string
}
