package options

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gomod.alauda.cn/alauda-backend/pkg/client"
	"gomod.alauda.cn/alauda-backend/pkg/server"
)

var (
	kubeConfigPathFlag *flag.Flag
)

const (
	flagAPIServerHost             = "apiserver-host"
	flagKubeConfigPath            = "kubeconfig"
	flagEnableAnonymous           = "enable-anonymous"
	flagQPS                       = "qps"
	flagBurst                     = "burst"
	flagUserAgent                 = "user-agent"
	flagClientTimeout             = "client-timeout"
	flagEnableMultiCluster        = "enable-multi-cluster"
	flagMultiClusterProxyHost     = "multi-cluster-host"
	flagMultiClusterParameterName = "cluster-param-name"
	flagEnableQueryToken          = "enable-query-token"
)

const (
	configAPIServerHost             = "client.apiserver_host"
	configKubeConfigPath            = "client.kubeconfig_path"
	configEnableAnonymous           = "client.enable_anonymous"
	configQPS                       = "client.qps"
	configBurst                     = "client.burst"
	configUserAgent                 = "client.user_agent"
	configClientTimeout             = "client.timeout"
	configEnableMultiCluster        = "client.enable_multi_cluster"
	configMultiClusterHost          = "client.multi_cluster_host"
	configMultiClusterParameterName = "client.cluster_param_name"
	configEnableQueryToken          = "client.enable-query-token"
)

// ClientOptions holds the options for client configuration.
type ClientOptions struct {
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

	// Enable multi-cluster proxy
	// if enabled will attempt to use the request parameter or query parameter cluster
	// as a cluster name used in the proxy
	EnableMultiCluster bool

	// If enabled the multi cluster host address must be specified
	MultiClusterHost string

	// MultiClusterParameterName parameter name for cluster generation
	// defaults to cluster
	MultiClusterParameterName string

	// EnableQueryToken allows users to provide authorization information with path query parameter
	// default query parameter name "token"
	EnableQueryToken bool
}

var _ Optioner = &ClientOptions{}

// NewClientOptions creates the default ClientOptions object.
func NewClientOptions() *ClientOptions {
	return &ClientOptions{
		EnableAnonymous:           false,
		Burst:                     1e6,
		QPS:                       1e6,
		UserAgent:                 "alauda-backend",
		Timeout:                   time.Second * 30,
		EnableMultiCluster:        false,
		MultiClusterParameterName: "cluster",
		EnableQueryToken:          false,
	}
}

// AddFlags adds flags related to debugging for controller manager to the specified FlagSet.
func (o *ClientOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.String(flagAPIServerHost, o.KubeAPIServer,
		"The address of the Kubernetes Apiserver "+
			"to connect to in the format of protocol://address:port, e.g., "+
			"http://localhost:8080. If not specified, the assumption is that the binary runs inside a "+
			"Kubernetes cluster and local discovery is attempted.")
	_ = viper.BindPFlag(configAPIServerHost, fs.Lookup(flagAPIServerHost))

	// "sigs.k8s.io/controller-runtime" will register "kubeconfig" flag automately in https://github.com/kubernetes-sigs/controller-runtime/blob/v0.8.3/pkg/client/config/config.go
	// if any import the package  "sigs.k8s.io/controller-runtime", will caused "kubeconfig" flag redefined.
	kubeConfigPathFlag := flag.Lookup(flagKubeConfigPath)
	if kubeConfigPathFlag == nil {
		fs.String(flagKubeConfigPath, o.KubeConfigPath,
			"Path to kubeconfig file with authorization and master location information.")
	}
	_ = viper.BindPFlag(configKubeConfigPath, fs.Lookup(flagKubeConfigPath))

	fs.String(flagUserAgent, o.UserAgent,
		"User agent used by the client")
	_ = viper.BindPFlag(configUserAgent, fs.Lookup(flagUserAgent))

	fs.Bool(flagEnableAnonymous, o.EnableAnonymous,
		"When enabled this settings will use the kubeconfig auth info or service account info instead of user login")
	_ = viper.BindPFlag(configEnableAnonymous, fs.Lookup(flagEnableAnonymous))

	fs.Float32(flagQPS, o.QPS,
		"QPS used by the client")
	_ = viper.BindPFlag(configQPS, fs.Lookup(flagQPS))

	fs.Int(flagBurst, o.Burst,
		"Burst used by the client")
	_ = viper.BindPFlag(configBurst, fs.Lookup(flagBurst))

	fs.Duration(flagClientTimeout, o.Timeout,
		"Timeout set on client")
	_ = viper.BindPFlag(configClientTimeout, fs.Lookup(flagClientTimeout))

	fs.Bool(flagEnableMultiCluster, o.EnableMultiCluster,
		"Enable multi-cluster client using request's cluster parameter name or query string. If true must set "+
			flagMultiClusterProxyHost+" as a full fledged hostname.")
	_ = viper.BindPFlag(configEnableMultiCluster, fs.Lookup(flagEnableMultiCluster))

	fs.String(flagMultiClusterProxyHost, o.MultiClusterHost,
		"Multi cluster host full fledged address.")
	_ = viper.BindPFlag(configMultiClusterHost, fs.Lookup(flagMultiClusterProxyHost))

	fs.String(flagMultiClusterParameterName, o.MultiClusterParameterName,
		"Multi cluster parameter name from request.")
	_ = viper.BindPFlag(configMultiClusterParameterName, fs.Lookup(flagMultiClusterParameterName))

	fs.Bool(flagEnableQueryToken, o.EnableQueryToken,
		"Enable query token client using request's token parameter name or query string.")
	_ = viper.BindPFlag(configEnableQueryToken, fs.Lookup(flagEnableQueryToken))
}

// ApplyFlags parsing parameters from the command line or configuration file
// to the options instance.
func (o *ClientOptions) ApplyFlags() []error {
	var errs []error

	if kubeConfigPathFlag != nil {
		o.KubeConfigPath = kubeConfigPathFlag.Value.String()
	} else {
		o.KubeConfigPath = viper.GetString(configKubeConfigPath)
	}

	o.KubeAPIServer = viper.GetString(configAPIServerHost)
	o.EnableAnonymous = viper.GetBool(configEnableAnonymous)
	o.QPS = float32(viper.GetFloat64(configQPS))
	o.Burst = viper.GetInt(configBurst)
	o.UserAgent = viper.GetString(configUserAgent)
	o.Timeout = viper.GetDuration(configClientTimeout)

	o.EnableMultiCluster = viper.GetBool(configEnableMultiCluster)
	o.MultiClusterHost = viper.GetString(configMultiClusterHost)
	o.MultiClusterParameterName = viper.GetString(configMultiClusterParameterName)

	o.EnableQueryToken = viper.GetBool(configEnableQueryToken)

	if o.EnableMultiCluster {
		if strings.TrimSpace(o.MultiClusterHost) == "" {
			errs = append(errs, fmt.Errorf(flagMultiClusterProxyHost+" must be set when "+flagEnableMultiCluster+" is enabled"))
		}
		if strings.TrimSpace(o.MultiClusterParameterName) == "" {
			errs = append(errs, fmt.Errorf(flagMultiClusterParameterName+" must be set when "+flagEnableMultiCluster+" is enabled"))
		}
	}

	return errs
}

// ApplyToServer apply options to server
func (o *ClientOptions) ApplyToServer(server server.Server) (err error) {
	if o == nil {
		return
	}
	mgr := client.NewManager().WithConfig(&client.Config{
		// EnableAnonymous: o.EnableAnonymous,
		KubeAPIServer:             o.KubeAPIServer,
		KubeConfigPath:            o.KubeConfigPath,
		QPS:                       o.QPS,
		Burst:                     o.Burst,
		UserAgent:                 o.UserAgent,
		Timeout:                   o.Timeout,
		MultiClusterHost:          o.MultiClusterHost,
		MultiClusterParameterName: o.MultiClusterParameterName,
		Log:                       server.L().Named("client-manager"),
	}).WithInsecure(client.InsecureConfigGenerator)

	if o.EnableMultiCluster {
		mgr.With(client.MultiClusterBearerTokenConfigGenerator)
	}

	mgr.With(client.BearerTokenConfigGenerator)

	if o.EnableAnonymous {
		mgr.With(client.InsecureConfigGenerator)
	}

	if o.EnableQueryToken {
		// mutiple cluster
		if o.EnableMultiCluster {
			mgr.With(client.QueryTokenMutipleClusterGenerator)
		}

		mgr.With(client.QueryTokenConfigGenerator)
	}

	server.SetManager(mgr)

	return
}
