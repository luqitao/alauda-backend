package options

import (
	"flag"

	"gomod.alauda.cn/alauda-backend/pkg/server"

	"github.com/spf13/pflag"
	"k8s.io/klog"
)

// KlogOptions options for klog
type KlogOptions struct {
}

var _ Optioner = &KlogOptions{}

// NewKlogOptions constructs klog options
func NewKlogOptions() *KlogOptions {
	return &KlogOptions{}
}

// AddFlags add flags to this option
func (o *KlogOptions) AddFlags(fs *pflag.FlagSet) {
	klog.InitFlags(nil)
	fs.AddGoFlagSet(flag.CommandLine)
}

// ApplyFlags parsing parameters from the command line or configuration file
// to the options instance.
func (o *KlogOptions) ApplyFlags() []error {
	return nil
}

// ApplyToServer apply options on server
func (o *KlogOptions) ApplyToServer(srv server.Server) error {
	return nil
}
