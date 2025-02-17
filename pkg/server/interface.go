package server

import (
	"net"

	"gomod.alauda.cn/alauda-backend/pkg/auth"

	"github.com/emicklei/go-restful/v3"
	"go.uber.org/zap"
	"gomod.alauda.cn/alauda-backend/pkg/audit"
	"gomod.alauda.cn/alauda-backend/pkg/client"
)

// Server interface for the server of this framework
type Server interface {
	Container() *restful.Container
	Start()

	Logger
	Manager
	ErrorHandler
	KeyValueGetterSetter
	AuditManager
	AuditWorker
	AuthManager
}

// ListenerSetter sets lister for server
type ListenerSetter interface {
	SetListener(net.Listener, int)
}

// PathPrefixSetter sets a path prefix
type PathPrefixSetter interface {
	SetPathPrefix(string)
}

// Logger interface to manage logger
type Logger interface {
	LoggerSetter
	LoggerGetter
}

// LoggerSetter sets a zap.Logger
type LoggerSetter interface {
	SetLogger(*zap.Logger)
}

// LoggerGetter gets a zap.Logger instance
type LoggerGetter interface {
	L() *zap.Logger
	Log() *zap.Logger
}

// Manager interface for setting a client manager into the server
type Manager interface {
	ManagerSetter
	ManagerGetter
}

// ManagerSetter sets a client.Manager
type ManagerSetter interface {
	SetManager(client.Manager)
}

// ManagerGetter interface to get a client.Manager
type ManagerGetter interface {
	GetManager() client.Manager
}

// ErrorHandler handle errors for requests
type ErrorHandler interface {
	HandleError(err error, req *restful.Request, res *restful.Response)
	ErrorHandlerSetter
}

// ErrorHandlerFunc error handling function
type ErrorHandlerFunc func(err error, req *restful.Request, res *restful.Response)

// ErrorHandlerSetter interface to set an error handler
type ErrorHandlerSetter interface {
	SetErrorHandler(ErrorHandlerFunc)
}

// KeyValueGetterSetter a interface for setting and getting values
type KeyValueGetterSetter interface {
	SetValue(key string, value interface{})
	GetValue(key string) (val interface{}, ok bool)
}

// AuditManager a interface for setting and getting audit.Manager
type AuditManager interface {
	SetAuditManager(audit.Manager)
	GetAuditManager() audit.Manager
}

// AuditWorker a interface fro setting and getting audit worker's number
type AuditWorker interface {
	SetAuditWorkerNum(int)
	GetAuditWorkerNum() int
	SetAuditQueueSize(int)
	GetAuditQueueSize() int
	EnqueueAuditJob(AuditJob)
}

// AuthManager a interface for setting and getting auth.Manager
type AuthManager interface {
	SetAuthManager(auth.Manager)
	GetAuthManager() auth.Manager
}
