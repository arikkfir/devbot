package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment represents a deployment of a repository into an environment.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +condition:Current,Stale:ApplyFailed,BakingFailed,CloneFailed,CloneMissing,Cloning,Pulling
// +condition:Current,Stale:InternalError,Invalid
// +condition:Valid,Invalid:AddFinalizerFailed,ControllerMissing,FailedGettingOwnedObjects,FinalizationFailed,InternalError
// +condition:Valid,Invalid:RepositoryNotAccessible,RepositoryNotFound,RepositoryNotSupported
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

	// Repository is the reference to the repository to be deployed.
	// +kubebuilder:validation:Required
	Repository NamespacedRepositoryReference `json:"repository,omitempty"`

	// Branch is the name of the branch to be deployed from the repository. This might be different from the branch
	// referenced in the owning Environment if the desired branch does not exist in the referenced
	// repository.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Branch string `json:"branch,omitempty"`
}

type DeploymentStatus struct {

	// Conditions represent the latest available observations of the deployment's state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

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
