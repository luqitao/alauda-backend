package options

import (
	"net/http"
	"net/http/pprof"
	"runtime"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gomod.alauda.cn/alauda-backend/pkg/server"
)

const (
	flagDebugProfiling           = "profiling"
	flagDebugContentionProfiling = "contention-profiling"
)

const (
	configDebugProfiling           = "debug.profiling"
	configDebugContentionProfiling = "debug.contention_profiling"
)

// DebugOptions holds the Debugging options.
type DebugOptions struct {
	EnableProfiling           bool
	EnableContentionProfiling bool
}

// NewDebugOptions creates the default DebugOptions object.
func NewDebugOptions() *DebugOptions {
	return &DebugOptions{
		EnableProfiling:           false,
		EnableContentionProfiling: false,
	}
}

// AddFlags adds flags related to debugging for controller manager to the specified FlagSet.
func (o *DebugOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.Bool(flagDebugProfiling, o.EnableProfiling,
		"Enable profiling via web interface host:port/debug/pprof/")
	_ = viper.BindPFlag(configDebugProfiling, fs.Lookup(flagDebugProfiling))
	fs.Bool(flagDebugContentionProfiling, o.EnableContentionProfiling,
		"Enable lock contention profiling, if profiling is enabled")
	_ = viper.BindPFlag(configDebugContentionProfiling, fs.Lookup(flagDebugContentionProfiling))
}

// ApplyFlags parsing parameters from the command line or configuration file
// to the options instance.
func (o *DebugOptions) ApplyFlags() []error {
	var errs []error

	o.EnableProfiling = viper.GetBool(configDebugProfiling)
	o.EnableContentionProfiling = viper.GetBool(configDebugContentionProfiling)

	return errs
}

// ApplyToServer apply options to server
func (o *DebugOptions) ApplyToServer(server server.Server) (err error) {
	if o == nil || !o.EnableProfiling {
		return
	}

	// setup.Debug{}.Install(server)
	server.Container().Handle("/debug/pprof", http.HandlerFunc(redirectTo("/debug/pprof/")))
	server.Container().Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	server.Container().Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	server.Container().Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	server.Container().Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	if o.EnableContentionProfiling {
		runtime.SetBlockProfileRate(1)
	}
	return
}

// redirectTo redirects request to a certain destination.
func redirectTo(to string) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		http.Redirect(rw, req, to, http.StatusFound)
	}
}
