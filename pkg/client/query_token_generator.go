package client

import (
	"fmt"

	restful "github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

// QueryTokenClientGenerator generates client based on request path token
func QueryTokenClientGenerator(cfg *Config, req *restful.Request) (client kubernetes.Interface, err error) {
	var config *rest.Config
	config, err = QueryTokenConfigGenerator(cfg, req)
	if err != nil {
		return
	}

	client, err = kubernetes.NewForConfig(config)
	return
}

// QueryTokenConfigGenerator returns a configuration given a Token
func QueryTokenConfigGenerator(cfg *Config, req *restful.Request) (config *rest.Config, err error) {
	config, err = cfg.Load()
	if err != nil {
		return
	}
	token := GetToken(req)
	// validates if the token is present and not empty, otherwise returns an error
	if token == "" {
		err = errors.NewUnauthorized("No Authorization Token provided")
		return
	}
	cmd := buildCmdConfig(&api.AuthInfo{Token: token}, config)
	config, err = cmd.ClientConfig()

	return
}

// QueryTokenMutipleClusterGenerator return config with cluster host and cluster name, if cluster name not set,just return config
func QueryTokenMutipleClusterGenerator(cfg *Config, req *restful.Request) (config *rest.Config, err error) {
	clusterName := GetClusterName(cfg.MultiClusterParameterName, req)
	if clusterName == "" {
		err = errors.NewBadRequest("Needs cluster parameter \"" + cfg.MultiClusterParameterName + "\"")
		return
	}
	config, err = QueryTokenConfigGenerator(cfg, req)
	if err != nil {
		return
	}

	if config != nil {
		config.Host = fmt.Sprintf("%s/kubernetes/%s", cfg.MultiClusterHost, clusterName)
	}
	return
}
