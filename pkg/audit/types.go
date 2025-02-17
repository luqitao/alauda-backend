package audit

import (
	authnv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	auditinternal "k8s.io/apiserver/pkg/apis/audit"
)

// Event captures all the information that can be included in an API audit log.
type Event struct {
	metav1.TypeMeta
	// AuditLevel at which event was generated
	Level auditinternal.Level
	// Unique audit ID, generated for each request.
	AuditID types.UID
	// Stage of the request handling when this event instance was generated.
	Stage auditinternal.Stage
	// RequestURI is the request URI as sent by the client to a server.
	RequestURI string
	// Verb is the advanced api verb associated with the request.
	Verb string
	// Authenticated user information.
	User authnv1.UserInfo
	// Impersonated user information.
	// +optional
	ImpersonatedUser *authnv1.UserInfo
	// Source IPs, from where the request originated and intermediate proxies.
	// +optional
	SourceIPs []string
	// UserAgent records the user agent string reported by the client.
	// Note that the UserAgent is provided by the client, and must not be trusted.
	// +optional
	UserAgent string
	// Object reference this request is targeted at.
	// Does not apply for List-type requests, or non-resource requests.
	// +optional
	ObjectRef *auditinternal.ObjectReference
	// The response status, populated even when the ResponseObject is not a Status type.
	// For successful responses, this will only include the Code. For non-status type
	// error responses, this will be auto-populated with the error Message.
	// +optional
	ResponseStatus *metav1.Status
	// API object from the request, in JSON format.
	// +optional
	RequestObject *map[string]interface{}
	// API object returned in the response, in JSON.
	// +optional
	ResponseObject *map[string]interface{}
	// Time the request reached the apiserver.
	RequestReceivedTimestamp metav1.MicroTime
	// Time the request reached current audit stage.
	StageTimestamp metav1.MicroTime
}

// Policy defines the configuration of audit logging
type Policy struct {
	metav1.TypeMeta

	// Rules is a list of PolicyRule, each request will only be processed by the first matched rule
	Rules []PolicyRule
}

// PolicyRule specify the request matching and processing methods
type PolicyRule struct {
	Level   auditinternal.Level `yaml:"level"`
	Match   RequestMatch        `yaml:"match,omitempty"`
	Process Process             `yaml:"process,omitempty"`
}

// RequestMatch defines rules used to match request
// only matched request will be processed
type RequestMatch struct {
	Path    string   `yaml:"path"`
	Methods []string `yaml:"methods"`
}

// Process defines rules to fullfill audit event object
type Process struct {
	VerbMatching map[string]string `yaml:"verbMatching,omitempty"`
	ObjectRef    ObjectReference   `yaml:"objectRef,omitempty"`
}

// ObjectReference specify how to generate Event.ObjectReference field
// Each filed can either be a template string or raw string
type ObjectReference struct {
	Resource    string `yaml:"resource"`
	Namespace   string `yaml:"namespace"`
	Name        string `yaml:"name"`
	APIGroup    string `yaml:"apiGroup"`
	APIVersion  string `yaml:"apiVersion"`
	SubResource string `yaml:"subResource"`
}
