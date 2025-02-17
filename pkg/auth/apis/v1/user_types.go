package v1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// State is User's State
type State string

// Action ...
type Action string

// ActionWithData ...
type ActionWithData struct {
	Action
	*Expired
}
type ConnectorData struct {
	Groups []string `json:"groups"`
}

const (
	UserKind            = "User"
	StateAcitve   State = "active"
	StateDisabled State = "disabled"
	StateLocked   State = "locked"
	StateInvalid  State = "invalid"
	StateDeleted  State = "deleted"

	ActionResetPassword     Action = "ActionResetPassword"
	ActionConfigValidPeriod Action = "expired"
	ActionActive            Action = "active"
	ActionDisable           Action = "disabled"
	ActionDelete            Action = "delete"
)

var (
	// ActionActiveToStateAcitve ..
	ActionActiveToStateAcitve = errors.NewBadRequest("User has been activated and cannot be operated")

	// ActionActiveToStateInvalid ..
	ActionActiveToStateInvalid = errors.NewBadRequest("User is invalid and inoperable")

	// ActionConfigValidePeriodToStateInvalid ..
	ActionConfigValidePeriodToStateInvalid = errors.NewBadRequest("User is invalid and inoperable")

	// ActionResetPasswordToStateInvalid ..
	ActionResetPasswordToStateInvalid = errors.NewBadRequest("User is invalid and inoperable")

	// ActionDisableToStateDisabled ..
	ActionDisableToStateDisabled = errors.NewBadRequest("User has been disabled and cannot be operated")

	// ActionToStateDeleted
	ActionToStateDeleted = errors.NewBadRequest("User was deleted and inoperable")
)

// UserSpec defines the desired state of User
type UserSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ConnectorType    string   `json:"connector_type"`
	ConnectorName    string   `json:"connector_name"`
	ConnectorId      string   `json:"connector_id,omitempty"`
	Email            string   `json:"email"`
	Username         string   `json:"username"`
	Groups           []string `json:"groups,omitempty"`
	Valid            bool     `json:"valid,omitempty"`
	IsAdmin          bool     `json:"is_admin"`
	Account          string   `json:"account,omitempty"`
	OldPassword      string   `json:"old_password,omitempty"`
	Password         string   `json:"password,omitempty"`
	ContinuityErrors int      `json:"continuity_errors,omitempty"`
	LastLoginTime    string   `json:"last_login_time,omitempty"`
	Expired          *Expired `json:"expired,omitempty"`
	State            State    `json:"state,omitempty"`
	IsDisabled       bool     `json:"is_disabled,omitempty"`
	Mail             string   `json:"mail,omitempty"`
	Mobile           string   `json:"mobile,omitempty"`
	WebhookUrl       string   `json:"webhookUrl,omitempty"`
	WebhookType      string   `json:"webhookType,omitempty"`
}

// Expired ...
type Expired struct {
	Begin mv1.Time `json:"begin"`
	End   mv1.Time `json:"end"`
}

// UserStatus defines the observed state of User
type UserStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

// User is the Schema for the users API
// +k8s:openapi-gen=true
type User struct {
	mv1.TypeMeta   `json:",inline"`
	mv1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec,omitempty"`
	Status UserStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// UserList contains a list of User
type UserList struct {
	mv1.TypeMeta `json:",inline"`
	mv1.ListMeta `json:"metadata,omitempty"`
	Items        []User `json:"items"`
}
