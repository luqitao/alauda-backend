package options

import (
	"fmt"

	"github.com/emicklei/go-restful/v3"
	"github.com/spf13/pflag"
	"gomod.alauda.cn/alauda-backend/pkg/server"
	"gomod.alauda.cn/log"
	"k8s.io/apimachinery/pkg/api/errors"
)

// ErrorOptions simple error options
type ErrorOptions struct{}

var _ Optioner = &ErrorOptions{}

// NewErrorOptions constructor for default error options
func NewErrorOptions() *ErrorOptions {
	return &ErrorOptions{}
}

// AddFlags no op for now
func (p *ErrorOptions) AddFlags(*pflag.FlagSet) {

}

// ApplyFlags no op for now
func (p *ErrorOptions) ApplyFlags() []error {
	return nil
}

// ApplyToServer sets a error handler for server
func (p *ErrorOptions) ApplyToServer(svr server.Server) error {
	svr.SetErrorHandler(func(err error, req *restful.Request, res *restful.Response) {
		statusErr := handleError(err)
		if statusErr.ErrStatus.APIVersion == "" {
			statusErr.ErrStatus.APIVersion = "v1"
		}
		if statusErr.ErrStatus.Kind == "" {
			statusErr.ErrStatus.Kind = "Status"
		}
		svr.L().Error("error handling request", log.Err(statusErr))
		res.WriteHeaderAndJson(int(statusErr.Status().Code), statusErr.Status(), restful.MIME_JSON)
	})
	return nil
}

func handleError(err error) *errors.StatusError {
	if err == nil {
		err = fmt.Errorf("Unknown error")
	}
	stats, ok := err.(*errors.StatusError)
	if !ok {
		stats = errors.NewInternalError(err)
	}
	return stats
}
