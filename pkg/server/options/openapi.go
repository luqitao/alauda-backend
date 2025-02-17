package options

import (
	"net/http"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/go-openapi/spec"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gomod.alauda.cn/alauda-backend/pkg/server"
	"gomod.alauda.cn/alauda-backend/pkg/server/options/assets/swagger"
)

// OpenAPIOptions provides the necessary to enable Open API capabilities to the server
type OpenAPIOptions struct {
	// EnableSwagger swagger document to be generated and enabled
	EnableSwagger bool

	// EnableSwaggerUI enables swagger UI
	EnableSwaggerUI bool

	// ObjectEnricherFunc provides a function to provide better api objects
	// on the swagger documentation
	ObjectEnricherFunc func(*spec.Swagger)
}

var _ Optioner = &OpenAPIOptions{}

const (
	flagOpenAPISwagger   = "swagger"
	flagOpenAPISwaggerUI = "swagger-ui"
)

const (
	configOpenAPISwagger   = "openapi.swagger"
	configOpenAPISwaggerUI = "openapi.swagger-ui"
)

// NewOpenAPIOptions creates options with defaults
func NewOpenAPIOptions() *OpenAPIOptions {
	return &OpenAPIOptions{
		EnableSwagger:   true,
		EnableSwaggerUI: true,
	}
}

// AddFlags adds flags for this option on the FlagSet.
func (o *OpenAPIOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.Bool(flagOpenAPISwagger, o.EnableSwagger,
		"Enable swagger api docs on /swagger.json")
	_ = viper.BindPFlag(configOpenAPISwagger, fs.Lookup(flagOpenAPISwagger))
	fs.Bool(flagOpenAPISwaggerUI, o.EnableSwaggerUI,
		"Enable swagger ui on /swagger-ui if swagger api docs is enabled")
	_ = viper.BindPFlag(configOpenAPISwaggerUI, fs.Lookup(flagOpenAPISwaggerUI))
}

// ApplyFlags parsing parameters from the command line or configuration file
// to the options instance.
func (o *OpenAPIOptions) ApplyFlags() []error {
	var err []error
	o.EnableSwagger = viper.GetBool(configOpenAPISwagger)
	o.EnableSwaggerUI = viper.GetBool(configOpenAPISwaggerUI)
	return err
}

// ApplyToServer apply options to server
func (o *OpenAPIOptions) ApplyToServer(server server.Server) (err error) {
	if o == nil || !o.EnableSwagger {
		return
	}

	config := restfulspec.Config{

		//WebServices:                   server.Container().RegisteredWebServices(),
		WebServices:                   server.Container().RegisteredWebServices(),
		APIPath:                       "/swagger.json",
		PostBuildSwaggerObjectHandler: o.ObjectEnricherFunc,
	}
	server.Container().Add(restfulspec.NewOpenAPIService(config))

	if o.EnableSwaggerUI {
		fileServer := http.FileServer(&assetfs.AssetFS{
			Asset:    swagger.Asset,
			AssetDir: swagger.AssetDir,
			Prefix:   "third_party/swagger-ui",
		})
		prefix := "/swagger-ui/"
		server.Container().Handle(prefix, http.StripPrefix(prefix, fileServer))
	}
	return
}
