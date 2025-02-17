package decorator

import (
	"bytes"
	"io/ioutil"
	"time"

	restful "github.com/emicklei/go-restful/v3"
	"gomod.alauda.cn/alauda-backend/pkg/audit"
	"gomod.alauda.cn/alauda-backend/pkg/httputil"
	"gomod.alauda.cn/alauda-backend/pkg/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Audit audit decorator. Used to generate, process and record audit logs from request and response
type Audit struct {
	server.Server
}

// NewAudit construtor for Audit decorator
func NewAudit(srv server.Server) Audit {
	return Audit{Server: srv}
}

// DefaultFilter filter to record audit logs according to policy file
func (a Audit) DefaultFilter(req *restful.Request, res *restful.Response, chain *restful.FilterChain) {
	mgr := a.GetAuditManager()
	requestReceivedTimestamp := metav1.NewMicroTime(time.Now())
	res.ResponseWriter = httputil.NewRespRecorderWriter(res.ResponseWriter)
	reqBody, _ := ioutil.ReadAll(req.Request.Body)
	req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
	chain.ProcessFilter(req, res)
	a.EnqueueAuditJob(&DefaultAuditJob{
		mgr:                      mgr,
		req:                      req,
		res:                      res,
		reqBody:                  &reqBody,
		requestReceivedTimestamp: requestReceivedTimestamp,
		handler:                  nil,
	})
}

// AuditHandler user defined handler used to fullfill audit event's fields
type AuditHandler func(*audit.Event, *restful.Request, *restful.Response)

// AuditJobGenerater for batch processing of multiple resources in a response
type AuditJobGenerater func(*restful.Request, *restful.Response, audit.Manager) []server.AuditJob

// NewCustomFilter returns a custom filer used to record audit logs according to user defined handler
func (a Audit) NewCustomFilter(handler AuditHandler) restful.FilterFunction {
	return func(req *restful.Request, res *restful.Response, chain *restful.FilterChain) {
		mgr := a.GetAuditManager()
		requestReceivedTimestamp := metav1.NewMicroTime(time.Now())
		_, ok := res.ResponseWriter.(*httputil.ResponseRecorderWriter)
		if !ok {
			res.ResponseWriter = httputil.NewRespRecorderWriter(res.ResponseWriter)
		}
		reqBody, _ := ioutil.ReadAll(req.Request.Body)
		req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
		chain.ProcessFilter(req, res)
		a.EnqueueAuditJob(&DefaultAuditJob{
			mgr:                      mgr,
			req:                      req,
			res:                      res,
			reqBody:                  &reqBody,
			requestReceivedTimestamp: requestReceivedTimestamp,
			handler:                  handler,
		})
	}
}

// NewBatchCustomFilter returns a custom filer used to record audit logs according to user defined handler
func (a Audit) NewBatchCustomFilter(generater AuditJobGenerater) restful.FilterFunction {
	return func(req *restful.Request, res *restful.Response, chain *restful.FilterChain) {
		chain.ProcessFilter(req, res)

		mgr := a.GetAuditManager()
		jobs := generater(req, res, mgr)
		for _, job := range jobs {
			a.EnqueueAuditJob(job)
		}
	}
}

// DefaultAuditJob default server.AuditJob implementation
type DefaultAuditJob struct {
	mgr                      audit.Manager
	req                      *restful.Request
	res                      *restful.Response
	reqBody                  *[]byte
	requestReceivedTimestamp metav1.MicroTime
	handler                  interface{}
}

// Execute generate and record audit event
func (aj *DefaultAuditJob) Execute() {
	aj.req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(*aj.reqBody))
	recorder, ok := aj.res.ResponseWriter.(*httputil.ResponseRecorderWriter)
	var resBodyBytes []byte
	if ok {
		resBodyBytes, _ = ioutil.ReadAll(recorder.Body)
		recorder.Body = bytes.NewBuffer(resBodyBytes)
	}
	ae := aj.mgr.NewAuditEvent(aj.requestReceivedTimestamp, aj.req.Request, aj.res.StatusCode(), bytes.NewBuffer(resBodyBytes))
	aj.mgr.ProcessUserInfo(ae, aj.req.Request)

	handlerFunc, isCustomHandler := aj.handler.(AuditHandler)
	if isCustomHandler {
		handlerFunc(ae, aj.req, aj.res)
	} else {
		matched, rule := aj.mgr.CheckIfRequestMatch(aj.req.Request)
		if !matched {
			return
		}
		aj.mgr.ExecutePolicyRule(ae, rule, aj.req.Request)
	}

	aj.mgr.Record(ae)
}
