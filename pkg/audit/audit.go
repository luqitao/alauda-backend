package audit

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	auditinternal "k8s.io/apiserver/pkg/apis/audit"
)

const (
	statusSuccess = "Success"
	statusFailure = "Failure"
)

// InitAuditEvent generates a audit event from request response
func InitAuditEvent(requestRecievedTimestamp metav1.MicroTime, req *http.Request, statusCode int, responseBody *bytes.Buffer) *Event {
	requestObject := make(map[string]interface{})
	responseObject := make(map[string]interface{})
	ev := &Event{
		Level:                    auditinternal.LevelRequestResponse,
		Stage:                    auditinternal.StageResponseComplete,
		RequestURI:               req.URL.RequestURI(),
		UserAgent:                maybeTruncateUserAgent(req),
		RequestObject:            &requestObject,
		ResponseObject:           &responseObject,
		RequestReceivedTimestamp: requestRecievedTimestamp,
		StageTimestamp:           metav1.NewMicroTime(time.Now()),
	}

	// prefer the id from the headers. If not available, create a new one.
	ids := req.Header.Get(auditinternal.HeaderAuditID)
	if ids != "" {
		ev.AuditID = types.UID(ids)
	} else {
		ev.AuditID = types.UID(uuid.New().String())
	}

	ips := utilnet.SourceIPs(req)
	ev.SourceIPs = make([]string, len(ips))
	for i := range ips {
		ev.SourceIPs[i] = ips[i].String()
	}

	reqBodyBytes, _ := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBodyBytes))
	decoder := json.NewDecoder(bytes.NewBuffer(reqBodyBytes))
	decoder.Decode(ev.RequestObject)

	status := statusSuccess
	if statusCode >= 400 {
		status = statusFailure
	}
	ev.ResponseStatus = &metav1.Status{
		Status: status,
		Code:   int32(statusCode),
	}

	if responseBody != nil {
		decoder := json.NewDecoder(responseBody)
		decoder.Decode(ev.ResponseObject)
	}
	return ev
}

// truncate User-Agent if too long, otherwise return it directly.
func maybeTruncateUserAgent(req *http.Request) string {
	ua := req.UserAgent()
	if len(ua) > maxUserAgentLength {
		ua = ua[:maxUserAgentLength] + userAgentTruncateSuffix
	}

	return ua
}
