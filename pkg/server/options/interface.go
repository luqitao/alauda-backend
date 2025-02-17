package options

import (
	"github.com/spf13/pflag"
	"gomod.alauda.cn/alauda-backend/pkg/server"
)

// Optioner interface for all options
type Optioner interface {
	AddFlags(*pflag.FlagSet)
	ApplyFlags() []error
	ApplyToServer(server.Server) error
}
