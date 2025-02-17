package client

import (
	restful "github.com/emicklei/go-restful/v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// InsecureClientGenFunc generates incluster service account insecure client
func InsecureClientGenFunc(cfg *Config, _ *restful.Request) (client kubernetes.Interface, err error) {
	var config *rest.Config
	config, err = InsecureConfigGenerator(cfg, nil)
	if err != nil {
		return
	}
	client, err = kubernetes.NewForConfig(config)
	return
}

// InsecureConfigGenerator generates a configuration for usage with insecure requests
// and will mainly use either a kube config data or a service account configuration
func InsecureConfigGenerator(cfg *Config, _ *restful.Request) (config *rest.Config, err error) {
	config, err = cfg.Load()
	return
}
