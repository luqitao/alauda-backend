package client

import (
	"fmt"

	restful "github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
)

// MultiClusterBearerTokenConfigGenerator configuration generator for multi-cluster client config
func MultiClusterBearerTokenConfigGenerator(cfg *Config, req *restful.Request) (config *rest.Config, err error) {
	clusterName := GetClusterName(cfg.MultiClusterParameterName, req)
	if clusterName == "" {
		err = errors.NewBadRequest("Needs cluster parameter \"" + cfg.MultiClusterParameterName + "\"")
		return
	}
	config, err = BearerTokenConfigGenerator(cfg, req)
	if config != nil {
		config.Host = fmt.Sprintf("%s/kubernetes/%s", cfg.MultiClusterHost, clusterName)
	}
	return
}
