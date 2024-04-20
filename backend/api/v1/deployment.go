package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment represents a deployment of a repository into an environment.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +condition:commons
// +condition:Current,Stale:InternalError,Invalid
// +condition:Current,Stale:PersistentVolumeCreationFailed,PersistentVolumeMissing
// +condition:Current,Stale:Cloning,CloneFailed,BranchNotFound,RepositoryNotAccessible,RepositoryNotFound
// +condition:Current,Stale:Baking,BakingFailed
// +condition:Current,Stale:Applying,ApplyFailed
// +condition:Valid,Invalid:RepositoryNotSupported
// +kubebuilder:printcolumn:name="Valid",type=string,JSONPath=`.status.privateArea.Valid`
// +kubebuilder:printcolumn:name="Repository",type=string,JSONPath=`.status.resolvedRepository`
// +kubebuilder:printcolumn:name="Branch",type=string,JSONPath=`.status.branch`
// +kubebuilder:printcolumn:name="PVC",type=string,JSONPath=`.status.persistentVolumeNameClaim`
// +kubebuilder:printcolumn:name="Last Attempted Revision",type=string,JSONPath=`.status.lastAttemptedRevision`
// +kubebuilder:printcolumn:name="Last Applied Revision",type=string,JSONPath=`.status.lastAppliedRevision`
// +kubebuilder:printcolumn:name="Current",type=string,JSONPath=`.status.privateArea.Current`
type Deployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the desired state of the Deployment.
	// +kubebuilder:validation:Required
	Spec DeploymentSpec `json:"spec"`

	// Status is the observed state of the Deployment.
	// +kubebuilder:validation:Optional
	Status DeploymentStatus `json:"status,omitempty"`
}

type DeploymentSpec struct {

	// Repository is the reference to the repository this deployment will deploy. The specific branch will be determined
	// during runtime based on existence of the parent environment's preferred branch and the repository's default
	// branch.
	// +kubebuilder:validation:Required
	Repository DeploymentRepositoryReference `json:"repository"`
}

type DeploymentStatus struct {

	// Conditions represent the latest available observations of the deployment's state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// The fully qualified (namespace + name) of the repository this deployment applies.
	// +kubebuilder:validation:Optional
	ResolvedRepository string `json:"resolvedRepository,omitempty"`

	// Branch is the actual branch being deployed from the repository. This may be the preferred branch from the parent
	// environment, or the repository's default branch if the preferred branch is not available.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Optional
	Branch string `json:"branch,omitempty"`

	// PersistentVolumeClaimName points to the name of the [k8s.io/api/core/v1.PersistentVolumeClaim] used for hosting
	// the cloned Git repository that this deployment will apply. The volume will be mounted to the various jobs this
	// deployment will create & run over its lifetime.
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-z0-9]+(\-[a-z0-9]+)*$
	// +kubebuilder:validation:Optional
	PersistentVolumeClaimName string `json:"persistentVolumeNameClaim,omitempty"`

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

	// PrivateArea is not meant for public consumption, nor is it part of the public API. It is exposed due to Go and
	// controller-runtime limitations but is an internal part of the implementation.
	PrivateArea ConditionsInverseState `json:"privateArea,omitempty"`
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
