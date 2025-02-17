package options

import (
	restful "github.com/emicklei/go-restful/v3"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gomod.alauda.cn/alauda-backend/pkg/server"
	"gomod.alauda.cn/log"
)

const (
	flagRequestLog = "request-log"
)

const (
	configRequestLog = "log.request_log"
)

// LogOptions options for logging
type LogOptions struct {
	*log.Options
	RequestLog bool
}

var _ Optioner = &LogOptions{}

// NewLogOptions constructs log options
func NewLogOptions() *LogOptions {
	return &LogOptions{
		Options:    log.NewOptions(),
		RequestLog: true,
	}
}

// AddFlags add flags to this option
func (o *LogOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	o.Options.AddFlags(fs)

	fs.Bool(flagRequestLog, o.RequestLog,
		"Enable request logs as debug")
	_ = viper.BindPFlag(configRequestLog, fs.Lookup(flagRequestLog))
}

// ApplyFlags apply flags to this option
func (o *LogOptions) ApplyFlags() []error {
	var errs []error

	errs = append(errs, o.Options.ApplyFlags()...)
	o.RequestLog = viper.GetBool(configRequestLog)

	return errs
}

// ApplyToServer apply to server
func (o *LogOptions) ApplyToServer(svr server.Server) (err error) {
	log.InitLogger(o.Options)
	if logSetter := svr.(server.LoggerSetter); logSetter != nil {
		logSetter.SetLogger(log.ZapLogger())
	}
	if o.RequestLog {
		svr.Container().Filter(func(req *restful.Request, res *restful.Response, chain *restful.FilterChain) {
			svr.L().Debug("==> request received", log.String("url", req.Request.RequestURI), log.Any("header", req.Request.Header))
			chain.ProcessFilter(req, res)
			svr.L().Debug("<== request resolved", log.String("url", req.Request.RequestURI), log.Int("status", res.StatusCode()), log.Any("header", req.Request.Header))
		})
	}
	return
}
