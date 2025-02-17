/*
Copyright 2018 Alauda.io.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ConditionType string
type UserBindingScope string

const (
	UserBindingScopePlatform  UserBindingScope = "platform"
	UserBindingScopeCluster   UserBindingScope = "cluster"
	UserBindingScopeProject   UserBindingScope = "project"
	UserBindingScopeNamespace UserBindingScope = "namespace"
)

type SubjectKind string

const (
	SubjectKindUser           = "User"
	SubjectKindGroup          = "Group"
	SubjectKindServiceAccount = "ServiceAccount"
)

type UserbindingsPhase string

type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type Subject struct {
	Kind SubjectKind `json:"kind"`
	Name string      `json:"name"`
}

type Constraint struct {
	Project   string `json:"project,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Cluster   string `json:"cluster,omitempty"`
}

type UserBindingCondition struct {
	LastTransitionTime *metav1.Time    `json:"lastTransitionTime,omitempty"`
	Message            string          `json:"message,omitempty"`
	Reason             string          `json:"reason,omitempty"`
	Status             ConditionStatus `json:"status"`
	Type               ConditionType   `json:"type"`
}

// UserBindingSpec defines the desired state of UserBinding
type UserBindingSpec struct {
	Subjects   []Subject        `json:"subjects"`
	RoleRef    string           `json:"roleRef"`
	Scope      UserBindingScope `json:"scope"`
	Constraint []Constraint     `json:"constraint,omitempty"`
}

// UserBindingStatus defines the observed state of UserBinding
type UserBindingStatus struct {
	Conditions   []UserBindingCondition `json:"conditions,omitempty"`
	LastSpecHash string                 `json:"lastSpecHash,omitempty"`
	Phase        UserbindingsPhase      `json:"phase,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UserBinding is the Schema for the userbindings API
// +k8s:openapi-gen=true
type UserBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserBindingSpec   `json:"spec,omitempty"`
	Status UserBindingStatus `json:"status,omitempty"`
}

func (u *UserBinding) RoleLevel() string {
	if level, ok := u.Labels["auth.cpaas.io/role.level"]; ok {
		return level
	}
	return ""
}

func (u *UserBinding) RoleName() string {
	if name, ok := u.Labels["auth.cpaas.io/role.name"]; ok {
		return name
	}
	return ""
}

func (u *UserBinding) UserEmail() string {
	if email, ok := u.Annotations["auth.cpaas.io/user.email"]; ok {
		return email
	}
	return ""
}

func (u *UserBinding) NamespaceCluster() string {
	if cluster, ok := u.Labels["cpaas.io/cluster"]; ok {
		return cluster
	}
	return ""
}

func (u *UserBinding) ConstraintClusters() []string {
	clusters := []string{}
	for _, c := range u.Spec.Constraint {
		clusters = append(clusters, c.Cluster)
	}
	return clusters
}

func (u *UserBinding) IsUserEmailExists() bool {
	if len(u.UserEmail()) == 0 {
		return false
	}
	return true
}

func (u *UserBinding) Email() string {
	if email, ok := u.Annotations["auth.cpaas.io/user.email"]; ok {
		return email
	}
	return ""
}

func (u *UserBinding) UserEmailName() string {
	if emailName, ok := u.Labels["auth.cpaas.io/user.email"]; ok {
		return emailName
	}
	return ""
}

func (u *UserBinding) CurrentCluster() string {
	if cluster, ok := u.Annotations["cpaas.io/current-cluster"]; ok {
		return cluster
	}
	return ""
}

func (u *UserBinding) IsCurrentClusterExists() bool {
	if len(u.CurrentCluster()) == 0 {
		return false
	}
	return true
}

func (u *UserBinding) ProjectName() string {
	if project, ok := u.Labels["cpaas.io/project"]; ok {
		return project
	}
	return ""
}

func (u *UserBinding) NamespaceName() string {
	if ns, ok := u.Labels["cpaas.io/namespace"]; ok {
		return ns
	}
	return ""
}

func (u *UserBinding) Subjects() []rbacv1.Subject {
	subjects := []rbacv1.Subject{}
	for _, s := range u.Spec.Subjects {
		subjects = append(subjects, rbacv1.Subject{
			APIGroup: rbacv1.GroupName,
			Kind:     string(s.Kind),
			Name:     s.Name,
		})
	}
	if len(subjects) == 0 {
		subjects = append(subjects, rbacv1.Subject{
			APIGroup: rbacv1.GroupName,
			Kind:     rbacv1.UserKind,
			Name:     u.Email(),
		})
	}
	return subjects
}

func (u *UserBinding) GroupName() string {
	if group, ok := u.Labels["auth.cpaas.io/group.name"]; ok {
		return group
	}
	return ""
}

func (u *UserBinding) SubjectKind() string {
	if u.Spec.Subjects != nil && len(u.Spec.Subjects) > 0 {
		return string(u.Spec.Subjects[0].Kind)
	}
	if _, ok := u.Labels["auth.cpaas.io/user.email"]; ok {
		return SubjectKindUser
	}
	if _, ok := u.Labels["auth.cpaas.io/group.name"]; ok {
		return SubjectKindGroup
	}
	return ""
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UserBindingList contains a list of UserBinding
type UserBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserBinding `json:"items"`
}
