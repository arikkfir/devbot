package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Environment is the Schema for the environments API.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +condition:Current,Stale:DeploymentsAreStale,InternalError,RepositoryNotAccessible,RepositoryNotFound,RepositoryNotReady,RepositoryNotSupported,DeploymentBranchOutOfSync,UnsupportedBranchStrategy
// +condition:Finalized,Finalizing:FinalizationFailed,FinalizerRemovalFailed,InProgress
// +condition:Initialized,FailedToInitialize:InternalError
// +condition:Valid,Invalid:ControllerNotAccessible,ControllerNotFound,ControllerReferenceMissing,InternalError
type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the desired state of the Environment.
	// +kubebuilder:validation:Required
	Spec EnvironmentSpec `json:"spec,omitempty"`

	// Status is the observed state of the Environment.
	// +kubebuilder:validation:Optional
	Status EnvironmentStatus `json:"status,omitempty"`
}

type EnvironmentSpec struct {
	// PreferredBranch is the preferred branch for deployment to this environment from each repository. Repositories
	// that lack this branch may opt to deploy their default branch instead (see [ApplicationSpecRepository.MissingBranchStrategy]).
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	PreferredBranch string `json:"branch,omitempty"`
}

type EnvironmentStatus struct {

	// Conditions represent the latest available observations of the application environment's state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true

type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Environment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Environment{}, &EnvironmentList{})
}
