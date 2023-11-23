package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +StatusCondition:GitHubRepositoryRef:Current

type GitHubRepositoryRefSpec struct {
	Ref string `json:"name,omitempty"`
}

type GitHubRepositoryRefStatus struct {
	CommitSHA  string             `json:"commitSHA,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type GitHubRepositoryRef struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitHubRepositoryRefSpec   `json:"spec,omitempty"`
	Status GitHubRepositoryRefStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type GitHubRepositoryRefList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitHubRepositoryRef `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitHubRepositoryRef{}, &GitHubRepositoryRefList{})
}
