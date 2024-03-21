package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment represents a deployment of a repository into an environment.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +condition:Current,Stale:Applying,ApplyFailed,Baking,BakingFailed,BranchNotFound,CheckoutFailed,CloneFailed,CloneMissing,CloneOpenFailed,Cloning,FetchFailed,InternalError,Invalid,RepositoryNotAccessible,RepositoryNotFound
// +condition:Finalized,Finalizing:FinalizationFailed,FinalizerRemovalFailed,InProgress
// +condition:Initialized,FailedToInitialize:InternalError
// +condition:Valid,Invalid:ControllerNotAccessible,ControllerNotFound,ControllerReferenceMissing,InternalError,RepositoryNotSupported
type Deployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the desired state of the Deployment.
	// +kubebuilder:validation:Required
	Spec DeploymentSpec `json:"spec,omitempty"`

	// Status is the observed state of the Deployment.
	// +kubebuilder:validation:Optional
	Status DeploymentStatus `json:"status,omitempty"`
}

type DeploymentSpec struct {

	// Repository is the reference to the repository this deployment will deploy. The specific branch will be determined
	// during runtime based on existence of the parent environment's preferred branch and the repository's default
	// branch.
	// +kubebuilder:validation:Required
	Repository NamespacedReference `json:"repository,omitempty"`
}

type DeploymentStatus struct {

	// Conditions represent the latest available observations of the deployment's state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// Branch is the actual branch being deployed from the repository. This may be the preferred branch from the parent
	// environment, or the repository's default branch if the preferred branch is not available.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Optional
	Branch string `json:"branch,omitempty"`

	// ClonePath is the local path in the controller to the cloned repository.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Optional
	ClonePath string `json:"clonePath,omitempty"`

	// LastAppliedCommitSHA is the commit SHA last applied (deployed) from the source into the target environment, if
	// any.
	// +kubebuilder:validation:MaxLength=40
	// +kubebuilder:validation:MinLength=40
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=^[a-f0-9]+$
	LastAttemptedRevision string `json:"lastAttemptedRevision,omitempty"`

	// LastAppliedCommitSHA is the commit SHA last applied (deployed) from the source into the target environment, if
	// any.
	// +kubebuilder:validation:MaxLength=40
	// +kubebuilder:validation:MinLength=40
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=^[a-f0-9]+$
	LastAppliedRevision string `json:"lastAppliedRevision,omitempty"`

	// Generates manifest of resources from the repository.
	// +kubebuilder:validation:Optional
	ResourcesManifest string `json:"resourcesManifest,omitempty"`
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
