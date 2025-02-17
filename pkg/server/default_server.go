package server

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"gomod.alauda.cn/alauda-backend/pkg/auth"

	"github.com/emicklei/go-restful/v3"
	"go.uber.org/zap"
	"gomod.alauda.cn/alauda-backend/pkg/audit"
	"gomod.alauda.cn/alauda-backend/pkg/client"
)

const (
	defaultWorkerNum = 15
	defaultQueueSize = 1000
)

// AuditJob interface for generating and recording audit event
type AuditJob interface {
	Execute()
}

// DefaultServer default implementation for server
type DefaultServer struct {
	container        *restful.Container
	listener         map[int]net.Listener
	listenerLock     *sync.RWMutex
	insecureBindPort int
	secureBindPort   int
	logger           *zap.Logger
	logLock          *sync.RWMutex
	manager          client.Manager
	mgrLock          *sync.RWMutex
	auditManager     audit.Manager
	auditWorkerNum   int
	auditQueueSize   int
	auditLock        *sync.RWMutex
	auditQueue       chan AuditJob
	authManager      auth.Manager
	authLock         *sync.RWMutex

	errHandler ErrorHandlerFunc

	store     map[string]interface{}
	storeLock *sync.RWMutex
}

var _ Server = &DefaultServer{}

// New initializes a new  default server
func New(name string) Server {
	logger, _ := zap.NewProduction()
	if logger == nil {
		logger = zap.NewNop()
	}
	server := &DefaultServer{
		container:    restful.NewContainer(),
		logger:       logger,
		logLock:      &sync.RWMutex{},
		listener:     map[int]net.Listener{},
		listenerLock: &sync.RWMutex{},
		mgrLock:      &sync.RWMutex{},
		auditLock:    &sync.RWMutex{},
		authLock:     &sync.RWMutex{},

		store:     map[string]interface{}{},
		storeLock: &sync.RWMutex{},
	}
	return server
}

// WithPort add custom port of server
func WithPort(svr Server, port int) Server {
	if server, ok := svr.(*DefaultServer); ok {
		server.insecureBindPort = port
		server.listener[port] = nil
	}
	return svr
}

// WithSecurePort add custom port of server
func WithSecurePort(svr Server, port int) Server {
	if server, ok := svr.(*DefaultServer); ok {
		server.secureBindPort = port
		server.listener[port] = nil
	}
	return svr
}

// Container returns container of server
func (s *DefaultServer) Container() *restful.Container {
	return s.container
}

// Start starts server
func (s *DefaultServer) Start() {
	// start audit workers on the background
	if s.auditWorkerNum <= 0 {
		s.auditWorkerNum = defaultWorkerNum
	}
	if s.auditQueueSize <= 0 {
		s.auditQueueSize = defaultQueueSize
	}
	queue := make(chan AuditJob, s.auditQueueSize)
	for i := 0; i < s.auditWorkerNum; i++ {
		go func() {
			for {
				job, ok := <-queue
				// if queue is closed, stop the worker
				if !ok {
					return
				}
				job.Execute()
			}
		}()
	}
	s.auditQueue = queue

	for port, listener := range s.listener {
		if port <= 0 {
			continue
		}
		if listener != nil {
			// TODO: give options to enable a more complex server
			// instead of using the default one
			go http.Serve(listener, s.Container())
		} else {
			go http.ListenAndServe(fmt.Sprintf(":%d", port), s.Container())
		}
	}
	select {}
}

var _ ListenerSetter = &DefaultServer{}

// SetListener sets a listener
func (s *DefaultServer) SetListener(listener net.Listener, port int) {
	s.listenerLock.Lock()
	defer s.listenerLock.Unlock()
	s.listener[port] = listener
}

var _ LoggerSetter = &DefaultServer{}

// SetLogger sets a zap.Logger instance on server
func (s *DefaultServer) SetLogger(logg *zap.Logger) {
	s.logLock.Lock()
	defer s.logLock.Unlock()
	s.logger = logg
}

// Log returns a zap.Logger
func (s *DefaultServer) Log() *zap.Logger {
	return s.L()
}

// L returns a zap.Logger
func (s *DefaultServer) L() *zap.Logger {
	s.logLock.RLock()
	defer s.logLock.RUnlock()
	return s.logger
}

// SetManager sets a client manager on the server
func (s *DefaultServer) SetManager(mgr client.Manager) {
	s.mgrLock.Lock()
	defer s.mgrLock.Unlock()
	s.manager = mgr
}

// GetManager gets a client manager from the server
func (s *DefaultServer) GetManager() client.Manager {
	s.mgrLock.RLock()
	defer s.mgrLock.RUnlock()
	return s.manager
}

// HandleError handle request errors
func (s *DefaultServer) HandleError(err error, req *restful.Request, res *restful.Response) {
	s.errHandler(err, req, res)
}

// SetErrorHandler sets error handler for request errors
func (s *DefaultServer) SetErrorHandler(errHandler ErrorHandlerFunc) {
	s.errHandler = errHandler
}

// SetValue sets one key and one value
func (s *DefaultServer) SetValue(key string, value interface{}) {
	s.storeLock.Lock()
	s.store[key] = value
	s.storeLock.Unlock()
}

// GetValue gets a value based on a key
func (s *DefaultServer) GetValue(key string) (val interface{}, ok bool) {
	s.storeLock.RLock()
	defer s.storeLock.RUnlock()
	val, ok = s.store[key]
	return
}

// SetAuditManager sets a audit manager on the server
func (s *DefaultServer) SetAuditManager(mgr audit.Manager) {
	s.auditLock.Lock()
	defer s.auditLock.Unlock()
	s.auditManager = mgr
}

// GetAuditManager gets audit manager from the server
func (s *DefaultServer) GetAuditManager() audit.Manager {
	s.auditLock.RLock()
	defer s.auditLock.RUnlock()
	return s.auditManager
}

// SetAuditWorkerNum sets the number of audit workers on the server
func (s *DefaultServer) SetAuditWorkerNum(num int) {
	s.auditLock.Lock()
	defer s.auditLock.Unlock()
	s.auditWorkerNum = num
}

// GetAuditWorkerNum gets audit worker's number from the server
func (s *DefaultServer) GetAuditWorkerNum() int {
	s.auditLock.RLock()
	defer s.auditLock.RUnlock()
	return s.auditWorkerNum
}

// EnqueueAuditJob send an audit job into audit queue. if queue is full, this operation will block.
func (s *DefaultServer) EnqueueAuditJob(job AuditJob) {
	s.auditQueue <- job
}

// SetAuditQueueSize sets the capacity of audit queue on the server
func (s *DefaultServer) SetAuditQueueSize(size int) {
	s.auditLock.Lock()
	defer s.auditLock.Unlock()
	s.auditQueueSize = size
}

// GetAuditQueueSize gets audit queue size from the server
func (s *DefaultServer) GetAuditQueueSize() int {
	s.auditLock.RLock()
	defer s.auditLock.RUnlock()
	return s.auditQueueSize
}

// SetAuthManager sets a auth manager on the server
func (s *DefaultServer) SetAuthManager(mgr auth.Manager) {
	s.authLock.Lock()
	defer s.authLock.Unlock()
	s.authManager = mgr
}

// GetAuthManager gets auth manager from the server
func (s *DefaultServer) GetAuthManager() auth.Manager {
	s.authLock.RLock()
	defer s.authLock.RUnlock()
	return s.authManager
}
