package options

import (
	"github.com/spf13/pflag"
	"gomod.alauda.cn/alauda-backend/pkg/registry"
	"gomod.alauda.cn/alauda-backend/pkg/server"
)

// APIRegistryOptions this is a non-op options
// its only function is to register all the api handlers
type APIRegistryOptions struct {
}

// NewAPIRegistryOptions constructor for devops handler options
func NewAPIRegistryOptions() *APIRegistryOptions {
	return &APIRegistryOptions{}
}

// AddFlags adds flags for log to the specified FlagSet object.
func (o *APIRegistryOptions) AddFlags(fs *pflag.FlagSet) {
}

// ApplyFlags parsing parameters from the command line or configuration file
// to the options instance.
func (o *APIRegistryOptions) ApplyFlags() []error {
	return nil
}

// ApplyToServer apply options on server
func (o *APIRegistryOptions) ApplyToServer(srv server.Server) error {
	services := registry.WebServices()
	if len(services) > 0 {
		for _, ws := range services {
			srv.Container().Add(ws)
		}
	}
	for path, httphandler := range registry.HTTPHandlers() {
		srv.Container().Handle(path, httphandler)
	}
	services, err := registry.Build(srv)
	if err != nil {
		return err
	}
	if len(services) > 0 {
		for _, ws := range services {
			srv.Container().Add(ws)
		}
	}
	return nil
}
