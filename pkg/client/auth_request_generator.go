package client

import (
	restful "github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (

	// UserConfigName configuration/context for user
	UserConfigName = "UserConfig"
)

// BearerTokenClientGenerator generates client based on Bearer token
func BearerTokenClientGenerator(cfg *Config, req *restful.Request) (client kubernetes.Interface, err error) {
	var config *rest.Config
	config, err = BearerTokenConfigGenerator(cfg, req)
	if err != nil {
		return
	}
	client, err = kubernetes.NewForConfig(config)
	return
}

// BearerTokenConfigGenerator returns a configuration given a Bearer Token
func BearerTokenConfigGenerator(cfg *Config, req *restful.Request) (config *rest.Config, err error) {
	config, err = cfg.Load()
	if err != nil {
		return
	}

	token := GetToken(req)
	// validates if the token is not empty, otherwise returns an error
	if token == "" {
		err = errors.NewUnauthorized("No Authorization Bearer Token provided")
		return
	}
	cmd := buildCmdConfig(&api.AuthInfo{Token: token}, config)
	config, err = cmd.ClientConfig()
	return
}

func buildCmdConfig(authInfo *api.AuthInfo, cfg *rest.Config) clientcmd.ClientConfig {
	cmdCfg := api.NewConfig()
	cmdCfg.Clusters[UserConfigName] = &api.Cluster{
		Server:                   cfg.Host,
		CertificateAuthority:     cfg.TLSClientConfig.CAFile,
		CertificateAuthorityData: cfg.TLSClientConfig.CAData,
		InsecureSkipTLSVerify:    cfg.TLSClientConfig.Insecure,
	}
	cmdCfg.AuthInfos[UserConfigName] = authInfo
	cmdCfg.Contexts[UserConfigName] = &api.Context{
		Cluster:  UserConfigName,
		AuthInfo: UserConfigName,
	}
	cmdCfg.CurrentContext = UserConfigName

	return clientcmd.NewDefaultClientConfig(
		*cmdCfg,
		&clientcmd.ConfigOverrides{},
	)
}
