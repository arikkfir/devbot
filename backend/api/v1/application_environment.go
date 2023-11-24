package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +StatusCondition:ApplicationEnvironment:Current

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type ApplicationEnvironment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationEnvironmentSpec   `json:"spec,omitempty"`
	Status ApplicationEnvironmentStatus `json:"status,omitempty"`
}

type ApplicationEnvironmentSpec struct {
	Branch string `json:"branch,omitempty"`
}

type ApplicationEnvironmentStatus struct {
	Conditions []metav1.Condition                `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
	Sources    []ApplicationEnvironmentStatusRef `json:"sources,omitempty"`
}

type ApplicationEnvironmentStatusRef struct {
	// Source is a reference to the correct GitHubRepositoryRef providing deployment manifests for this environment.
	// It's possible that for an environment called "myBranch", one of the sources would be a pointer to a
	// GitHubRepositoryRef of a branch "main", because "myBranch" does not exist in the source repository.
	Source corev1.ObjectReference `json:"source,omitempty"`

	// LastAppliedCommitSHA is the commit SHA last applied (deployed) from the source into the target environment (can
	// be empty, if never deployed).
	LastAppliedCommitSHA string `json:"lastAppliedCommitSHA,omitempty"`
}

//+kubebuilder:object:root=true

type ApplicationEnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationEnvironment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApplicationEnvironment{}, &ApplicationEnvironmentList{})
}
