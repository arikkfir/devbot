package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Environment is the Schema for the environments API.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +condition:Current,Stale:InternalError,Invalid
// +condition:Current,Stale:RepositoryAdded,RepositoryMissingDefaultBranch,RepositoryNotAccessible,RepositoryNotApplicable,RepositoryNotFound,RepositoryNotSupported
// +condition:Valid,Invalid:AddFinalizerFailed,ControllerMissing,FailedGettingOwnedObjects,FinalizationFailed,InternalError
// +condition:Valid,Invalid:DeploymentNotFound,RepositoryNotSupported
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

	// Sources is the concrete set of branches to deploy to this environment from each participating repository. The
	// set of repositories participating is the repositories mentioned in the [ApplicationSpec.Repositories] that either
	// have a branch matching [EnvironmentSpec.PreferredBranch] or that their [ApplicationSpecRepository.MissingBranchStrategy]
	// is set to "UseDefaultBranch", in which case their default branch will be listed here.
	// +kubebuilder:validation:Required
	Sources []EnvironmentStatusSource `json:"branches,omitempty"`
}

type EnvironmentStatusSource struct {
	// Repository to be deployed to this environment.
	// +kubebuilder:validation:Required
	Repository NamespacedRepositoryReference `json:"repository,omitempty"`

	// Deployment refers to the deployment responsible for deploying this branch & repository to this environment.
	// +kubebuilder:validation:Optional
	Deployment *DeploymentReference `json:"deployment,omitempty"`
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
