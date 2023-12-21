package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +StatusCondition:Deployment:Deploying
// +StatusCondition:Deployment:Current

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type Deployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentSpec   `json:"spec,omitempty"`
	Status DeploymentStatus `json:"status,omitempty"`
}

type DeploymentSpec struct {
	Repository corev1.ObjectReference `json:"repository,omitempty"`
	Branch     string                 `json:"branch,omitempty"`
}

type DeploymentStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// LastAppliedCommitSHA is the commit SHA last applied (deployed) from the source into the target environment (can
	// be empty, if never deployed).
	LastAppliedCommitSHA string `json:"lastAppliedCommitSHA,omitempty"`
}

//+kubebuilder:object:root=true

type DeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Deployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Deployment{}, &DeploymentList{})
}
