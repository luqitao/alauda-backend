package options

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gomod.alauda.cn/alauda-backend/pkg/audit"
	"gomod.alauda.cn/alauda-backend/pkg/server"
)

const (
	flagAuditLogPath      = "audit-log-path"
	flagAuditLogMaxSize   = "audit-log-maxsize"
	flagAuditLogMaxBackup = "audit-log-maxbackup"
	flagAuditPolicyFile   = "audit-policy-file"
	flagAuditWorkerNum    = "audit-worker-num"
	flagAuditQueueSize    = "audit-queue-size"
)

const (
	configAuditLogPath      = "audit.log_path"
	configAuditLogMaxSize   = "audit.log_maxsize"
	configAuditLogMaxBackup = "audit.log_maxbackup"
	configAuditPolicyFile   = "audit.policy_file"
	configAuditWorkerNum    = "audit.worker_num"
	configAuditQueueSize    = "audit.queue_size"
)

// AuditOptions holds the options for audit configuration.
type AuditOptions struct {
	// If set, all requests coming to the apiserver will be logged to this file.
	// '-' means standard out.
	LogPath string
	// The maximum size in megabytes of the audit log file before it gets rotated.
	LogMaxSize int
	// The maximum number of old audit log files to retain.
	LogMaxBackup int
	// Path to the file that defines the audit policy configuration.
	PolicyFile string
	// The number of audit workers to record audit events.
	WorkerNum int
	// The size of audit queue(cache)
	QueueSize int
}

var _ Optioner = &ClientOptions{}

// NewAuditOptions creates the default AuditOptions object.
func NewAuditOptions() *AuditOptions {
	return &AuditOptions{
		LogMaxSize:   200,
		LogMaxBackup: 2,
		PolicyFile:   "/etc/audit/policy.yaml",
		WorkerNum:    15,
		QueueSize:    1000,
	}
}

// AddFlags adds flags related to audit for controller manager to the specified FlagSet.
func (o *AuditOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.String(flagAuditLogPath, o.LogPath,
		"If set, all requests coming to the apiserver will be "+
			"logged to this file. '-' means standard out.")
	_ = viper.BindPFlag(configAuditLogPath, fs.Lookup(flagAuditLogPath))

	fs.Int(flagAuditLogMaxSize, o.LogMaxSize,
		"The maximum size in megabytes of the audit log file before it gets rotated.")
	_ = viper.BindPFlag(configAuditLogMaxSize, fs.Lookup(flagAuditLogMaxSize))

	fs.Int(flagAuditLogMaxBackup, o.LogMaxBackup,
		"The maximum number of old audit log files to retain.")
	_ = viper.BindPFlag(configAuditLogMaxBackup, fs.Lookup(flagAuditLogMaxBackup))

	fs.String(flagAuditPolicyFile, o.PolicyFile,
		"Path to the file that defines the audit policy configuration.")
	_ = viper.BindPFlag(configAuditPolicyFile, fs.Lookup(flagAuditPolicyFile))

	fs.Int(flagAuditWorkerNum, o.WorkerNum,
		"The number of audit workers to record audit events.")
	_ = viper.BindPFlag(configAuditWorkerNum, fs.Lookup(flagAuditWorkerNum))

	fs.Int(flagAuditQueueSize, o.QueueSize,
		"The size of audit job queue.")
	_ = viper.BindPFlag(configAuditQueueSize, fs.Lookup(flagAuditQueueSize))
}

// ApplyFlags parsing parameters from the command line or configuration file
// to the options instance.
func (o *AuditOptions) ApplyFlags() []error {
	var errs []error

	o.LogPath = viper.GetString(configAuditLogPath)
	o.LogMaxSize = viper.GetInt(configAuditLogMaxSize)
	o.LogMaxBackup = viper.GetInt(configAuditLogMaxBackup)
	o.PolicyFile = viper.GetString(configAuditPolicyFile)
	o.WorkerNum = viper.GetInt(configAuditWorkerNum)
	o.QueueSize = viper.GetInt(configAuditQueueSize)

	if _, err := audit.LoadPolicyFromFile(o.PolicyFile); err != nil {
		errs = append(errs, fmt.Errorf("audit policy file invalid: %v", err.Error()))
	}

	return errs
}

// ApplyToServer apply options to server
func (o *AuditOptions) ApplyToServer(server server.Server) (err error) {
	if o == nil {
		return
	}
	mgr := audit.NewManager(&audit.Config{
		PolicyPath:    o.PolicyFile,
		LogPath:       o.LogPath,
		LogMaxSize:    o.LogMaxSize,
		LogMaxBackups: o.LogMaxBackup,
	})

	server.SetAuditManager(mgr)
	server.SetAuditWorkerNum(o.WorkerNum)
	server.SetAuditQueueSize(o.QueueSize)

	return
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
