package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +condition:Current,Stale:

// GitHubRepositoryRef is the Schema for the githubrepositoryrefs API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +condition:Authenticated,Unauthenticated:AuthSecretForbidden,AuthSecretGetFailed,AuthSecretKeyNotFound,AuthSecretNotFound,AuthTokenEmpty,Invalid,TokenValidationFailed
// +condition:Current,Stale:CommitSHAOutOfSync,Invalid,InternalError,RepositoryNameOutOfSync,RepositoryNotFound,RepositoryOwnerOutOfSync,Unauthenticated
// +condition:Finalized,Finalizing:FinalizationFailed,FinalizerRemovalFailed,InProgress
// +condition:Initialized,FailedToInitialize:InternalError
// +condition:Valid,Invalid:AuthConfigMissing,AuthSecretKeyMissing,AuthSecretNameMissing,ControllerNotAccessible,ControllerNotFound,ControllerReferenceMissing,InternalError,InvalidRefreshInterval,RepositoryNameMissing,RepositoryOwnerMissing
type GitHubRepositoryRef struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the desired state of the GitHubRepositoryRef.
	// +kubebuilder:validation:Required
	Spec GitHubRepositoryRefSpec `json:"spec,omitempty"`

	// Status is the observed state of the GitHubRepositoryRef.
	// +kubebuilder:validation:Optional
	Status GitHubRepositoryRefStatus `json:"status,omitempty"`
}

type GitHubRepositoryRefSpec struct {

	// Ref is the Git reference that this object represents. In Git, refs are represented usually as "refs/heads/branch"
	// or "refs/tags/tag". This field is used to identify the object in the Git repository.
	// +kubebuilder:validation:MaxLength=1000
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-zA-Z0-9][a-zA-Z0-9._-]*(/[a-zA-Z0-9][a-zA-Z0-9._-]*)*$
	// +kubebuilder:validation:Required
	Ref string `json:"ref,omitempty"`
}

type GitHubRepositoryRefStatus struct {

	// CommitSHA is the latest commit SHA of this Git reference.
	// +kubebuilder:validation:MaxLength=40
	// +kubebuilder:validation:MinLength=40
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=^[a-f0-9]+$
	CommitSHA string `json:"commitSHA,omitempty"`

	// Conditions represent the latest available observations of the GitHubRepositoryRef's state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// Owner is the owner of the GitHub repository that owns this ref.
	// +kubebuilder:validation:MaxLength=39
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=^[a-zA-Z0-9][a-zA-Z0-9-_]*[a-zA-Z0-9_]$
	RepositoryOwner string `json:"repositoryOwner,omitempty"`

	// Owner is the owner of the GitHub repository that owns this ref.
	// +kubebuilder:validation:MaxLength=100
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=^[a-zA-Z0-9_][a-zA-Z0-9-_.]*[a-zA-Z0-9_.]$
	RepositoryName string `json:"repositoryName,omitempty"`
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
