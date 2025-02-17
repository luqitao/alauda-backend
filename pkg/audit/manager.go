package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/natefinch/lumberjack"
	authnv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/authenticator"
)

const (
	maxUserAgentLength      = 1024
	userAgentTruncateSuffix = "...TRUNCATED"
	anonymousUser           = "system:anonymous"
)

// DefaultManager default audit manager
type DefaultManager struct {
	recorder    io.Writer
	policy      *Policy
	encoder     *json.Encoder
	tokenParser authenticator.Token
}

// Config is used to generate a DefaultManager
type Config struct {
	PolicyPath    string
	LogPath       string
	LogMaxSize    int
	LogMaxBackups int
}

// NewManager creates a DefaultManager instance
func NewManager(config *Config) *DefaultManager {
	policy, _ := LoadPolicyFromFile(config.PolicyPath)
	recorder := &lumberjack.Logger{
		Filename:   config.LogPath,
		MaxSize:    config.LogMaxSize,
		MaxBackups: config.LogMaxBackups,
	}
	return &DefaultManager{
		recorder:    recorder,
		policy:      policy,
		encoder:     json.NewEncoder(recorder),
		tokenParser: NewOIDCTokenParser(),
	}
}

// NewAuditEvent create and initialize a audit event object
func (mgr *DefaultManager) NewAuditEvent(requestRecievedTimestamp metav1.MicroTime, req *http.Request, statusCode int, responseBody *bytes.Buffer) *Event {
	return InitAuditEvent(requestRecievedTimestamp, req, statusCode, responseBody)
}

// ProcessUserInfo will fullfill audit event's User filed
func (mgr *DefaultManager) ProcessUserInfo(ae *Event, req *http.Request) {
	ae.User = authnv1.UserInfo{
		Username: anonymousUser,
	}

	token := GetToken(req)
	if token == "" {
		return
	}

	userResp, _, err := mgr.tokenParser.AuthenticateToken(context.TODO(), token)
	if err != nil {
		return
	}
	ae.User = authnv1.UserInfo{
		Username: userResp.User.GetName(),
		UID:      userResp.User.GetUID(),
		Groups:   userResp.User.GetGroups(),
	}
}

// CheckIfRequestMatch checks if request match the policy rules
func (mgr *DefaultManager) CheckIfRequestMatch(req *http.Request) (matched bool, rule PolicyRule) {
	return CheckPolicyMatch(req, mgr.policy)
}

// ExecutePolicyRule will fullfill audit event's Verb and ObjectRef fields according to policy rules
func (mgr *DefaultManager) ExecutePolicyRule(e *Event, r PolicyRule, req *http.Request) {
	ExecutePolicyProcess(e, r, req)
}

// Record will log audit event to specified audit log file
func (mgr *DefaultManager) Record(ae *Event) error {
	return mgr.encoder.Encode(ae)
}
