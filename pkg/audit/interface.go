package audit

import (
	"bytes"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Manager can be used to create, process and record audit events
type Manager interface {
	CheckIfRequestMatch(req *http.Request) (matched bool, rule PolicyRule)
	NewAuditEvent(requestRecievedTimestamp metav1.MicroTime, req *http.Request, statusCode int, responseBody *bytes.Buffer) *Event
	ProcessUserInfo(*Event, *http.Request)
	ExecutePolicyRule(*Event, PolicyRule, *http.Request)
	Record(*Event) error
}
