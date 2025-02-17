package context

import (
	"context"

	"go.uber.org/zap"
	"gomod.alauda.cn/alauda-backend/pkg/dataselect"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type contextKey struct{ Name string }

var (
	kubernetesClientKey = contextKey{Name: "kubernetes.Interface"}
	dynamicClientKey    = contextKey{Name: "dynamic.NamespaceableResourceInterface"}
	loggerKey           = contextKey{Name: "zap.Logger"}
	dataselectQueryKey  = contextKey{Name: "dataselect.Query"}
)

// WithClient inserts a client into the context
func WithClient(ctx context.Context, client kubernetes.Interface) context.Context {
	return context.WithValue(ctx, kubernetesClientKey, client)
}

// Client fetches a client from a context if existing.
// will return nil if the context doesnot have the client
func Client(ctx context.Context) kubernetes.Interface {
	val := ctx.Value(kubernetesClientKey)
	if val != nil {
		return val.(kubernetes.Interface)
	}
	return nil
}

// WithLogger writes a logger to a context without any addition
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// Logger returns a logger from a context. Will return nil if
// the context does not have the logger
func Logger(ctx context.Context) *zap.Logger {
	val := ctx.Value(loggerKey)
	if val != nil {
		return val.(*zap.Logger)
	}
	return nil
}

// WithDynamicClient inserts a DynamicClient into the context
func WithDynamicClient(ctx context.Context, client dynamic.NamespaceableResourceInterface) context.Context {
	return context.WithValue(ctx, dynamicClientKey, client)
}

// DynamicClient fetches a client from a context if existing.
// will return nil if the context doesnot have the client
func DynamicClient(ctx context.Context) dynamic.NamespaceableResourceInterface {
	val := ctx.Value(dynamicClientKey)
	if val != nil {
		return val.(dynamic.NamespaceableResourceInterface)
	}
	return nil
}

// WithQuery inserts a *dataselect.Query into the context
func WithQuery(ctx context.Context, selector *dataselect.Query) context.Context {
	return context.WithValue(ctx, dataselectQueryKey, selector)
}

// Query fetches a *dataselect.Query from a context if existing.
// will return nil if the context doesnot have the value
func Query(ctx context.Context) *dataselect.Query {
	val := ctx.Value(dataselectQueryKey)
	if val != nil {
		return val.(*dataselect.Query)
	}
	return nil
}
