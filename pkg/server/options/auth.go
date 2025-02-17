package options

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gomod.alauda.cn/alauda-backend/pkg/auth"
	"gomod.alauda.cn/alauda-backend/pkg/auth/clusterrole"
	"gomod.alauda.cn/alauda-backend/pkg/auth/request"
	"gomod.alauda.cn/alauda-backend/pkg/auth/user"
	"gomod.alauda.cn/alauda-backend/pkg/auth/userbinding"
	"gomod.alauda.cn/alauda-backend/pkg/server"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type AuthOptions struct {
	UserCacheExpire time.Duration
	APIPrefixes     []string
	SystemNamespace string
	ErebusService   string //
}

var _ Optioner = &AuthOptions{}

func NewAuthOptions() *AuthOptions {
	return &AuthOptions{}
}

// AddFlags adds flags related to audit for controller manager to the specified FlagSet.
func (o *AuthOptions) AddFlags(fs *pflag.FlagSet) {

}

// ApplyFlags parsing parameters from the command line or configuration file
// to the options instance.
func (o *AuthOptions) ApplyFlags() []error {
	o.ErebusService = viper.GetString("KUBERNETES_SERVICE_HOST")

	return nil
}

// ApplyToServer apply options to server
func (o *AuthOptions) ApplyToServer(server server.Server) (err error) {
	if o == nil {
		return
	}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	cache := auth.NewCache(1 * time.Minute)

	requestInfoResolver := &request.RequestInfoFactory{
		APIPrefixes: sets.NewString("platform", ""),
	}

	stopCh := make(chan struct{})
	userbindingResolver := userbinding.NewResolver(clientset, stopCh)
	clusterroleResolver := clusterrole.NewResolver(clientset, stopCh)
	userResolver := user.NewResolver(stopCh)

	mgr := auth.NewManager(cache, o.ErebusService, userbindingResolver, clusterroleResolver, requestInfoResolver, userResolver)
	server.SetAuthManager(mgr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigCh
		close(stopCh)
	}()
	return nil
}
