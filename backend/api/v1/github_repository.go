package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +StatusCondition:GitHubRepository:AuthenticatedToGitHub
// +StatusCondition:GitHubRepository:Current

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type GitHubRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitHubRepositorySpec   `json:"spec,omitempty"`
	Status GitHubRepositoryStatus `json:"status,omitempty"`
}

type GitHubRepositorySpec struct {
	Owner string               `json:"owner,omitempty"`
	Name  string               `json:"name,omitempty"`
	Auth  GitHubRepositoryAuth `json:"auth,omitempty"`
}

type GitHubRepositoryAuth struct {
	PersonalAccessToken *GitHubRepositoryAuthPersonalAccessToken `json:"personalAccessToken,omitempty"`
}

type GitHubRepositoryAuthPersonalAccessToken struct {
	Secret corev1.SecretReference `json:"secret,omitempty"`
	Key    string                 `json:"key,omitempty"`
}

type GitHubRepositoryStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true

type GitHubRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitHubRepository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitHubRepository{}, &GitHubRepositoryList{})
}
