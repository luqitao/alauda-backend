package client

import (
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Config configuration options for client generation
type Config struct {
	// KubeAPIServer full address for the kubernetes apiserver
	// if not provided, will use inCluster configuration
	KubeAPIServer string

	// KubeConfigPath path to kubeconfig file
	// used mostly for development environment
	// or out of cluster deployment
	KubeConfigPath string

	// EnableAnonymous allows users to not provide authorization information and
	// uses service account information or kube config authorization
	EnableAnonymous bool

	// QPS indicates the maximum QPS to the master from this client.
	// If it's zero, the created RESTClient will use DefaultQPS: 5
	QPS float32

	// Maximum burst for throttle.
	// If it's zero, the created RESTClient will use DefaultBurst: 10.
	Burst int

	// UserAgent is an optional field that specifies the caller of this request.
	UserAgent string

	// The maximum length of time to wait before giving up on a server request. A value of zero means no timeout.
	Timeout time.Duration

	// AcceptContentTypes specifies the types the client will accept and is optional.
	// If not set, ContentType will be used to define the Accept header
	AcceptContentTypes string
	// ContentType specifies the wire format used to communicate with the server.
	// This value will be set as the Accept header on requests made to the server, and
	// as the default content type on any object sent to the server. If not set,
	// "application/json" is used.
	ContentType string

	// MultiClusterHost specifies the host address for the multi-cluster proxy
	MultiClusterHost string

	// MultiClusterParameterName parameter name used to fetch multi cluster data
	MultiClusterParameterName string

	// Logger instance
	Log *zap.Logger
}

// Load loads configuration
func (g *Config) Load() (config *rest.Config, err error) {
	defer func() {
		g.setupConfig(config, err)
	}()
	if g.KubeAPIServer == "" && g.KubeConfigPath == "" {
		config, err = rest.InClusterConfig()
	}
	if config != nil {
		return
	}
	config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: g.KubeConfigPath},
		&clientcmd.ConfigOverrides{ClusterInfo: api.Cluster{Server: g.KubeAPIServer, InsecureSkipTLSVerify: true}}).ClientConfig()
	return
}

func (g *Config) setupConfig(cfg *rest.Config, err error) {
	if err == nil && cfg != nil {
		cfg.Burst = g.Burst
		cfg.QPS = g.QPS
		cfg.UserAgent = g.UserAgent
		cfg.ContentType = g.ContentType
		cfg.Timeout = g.Timeout
		cfg.AcceptContentTypes = g.AcceptContentTypes

		// skip secure check and ingore CA
		cfg.TLSClientConfig = rest.TLSClientConfig{
			Insecure: true,
			CAFile:   "",
			CAData:   nil,
		}
	}
}
