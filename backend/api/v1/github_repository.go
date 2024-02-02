package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GitHubRepository is the Schema for the githubrepositories API.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +condition:Authenticated,Unauthenticated:AuthSecretForbidden,AuthSecretGetFailed,AuthSecretKeyNotFound,AuthSecretNotFound,AuthTokenEmpty,Invalid,TokenValidationFailed
// +condition:Current,Stale:BranchesOutOfSync,DefaultBranchOutOfSync,GitHubAPIFailure,InternalError,Invalid,RepositoryNotFound,Unauthenticated
// +condition:Finalized,Finalizing:FinalizationFailed,FinalizerRemovalFailed,InProgress
// +condition:Initialized,FailedToInitialize:InternalError
// +condition:Valid,Invalid:AuthConfigMissing,AuthSecretKeyMissing,AuthSecretNameMissing,InvalidRefreshInterval,RepositoryNameMissing,RepositoryOwnerMissing
type GitHubRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the desired state of the GitHubRepository.
	// +kubebuilder:validation:Required
	Spec GitHubRepositorySpec `json:"spec,omitempty"`

	// Status is the observed state of the GitHubRepository.
	// +kubebuilder:validation:Optional
	Status GitHubRepositoryStatus `json:"status,omitempty"`
}

type GitHubRepositorySpec struct {

	// Owner is the GitHub user or organization that owns the repository.
	// +kubebuilder:validation:MaxLength=39
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-zA-Z0-9][a-zA-Z0-9-_]*[a-zA-Z0-9_]$
	// +kubebuilder:validation:Required
	Owner string `json:"owner,omitempty"`

	// Name is the name of the repository.
	// +kubebuilder:validation:MaxLength=100
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-zA-Z0-9_][a-zA-Z0-9-_.]*[a-zA-Z0-9_.]$
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// Auth is the authentication method to use when accessing the repository.
	// +kubebuilder:validation:Required
	Auth GitHubRepositoryAuth `json:"auth,omitempty"`

	// RefreshInterval is the interval at which to refresh the list of branches in the repository. The value should be
	// specified as a duration string, e.g. "5m" for 5 minutes. The default value is "5m".
	// +kubebuilder:default="5m"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Optional
	RefreshInterval string `json:"refreshInterval,omitempty"`
}

type GitHubRepositoryAuth struct {

	// PersonalAccessToken signals that we should use a GitHub personal access token (PAT) when accessing the repository
	// as well as how to obtain it.
	// +kubebuilder:validation:Optional
	PersonalAccessToken *GitHubRepositoryAuthPersonalAccessToken `json:"personalAccessToken,omitempty"`
}

type GitHubRepositoryAuthPersonalAccessToken struct {

	// Secret is the reference to the secret containing the GitHub personal access token.
	// +kubebuilder:validation:Required
	Secret SecretReferenceWithOptionalNamespace `json:"secret,omitempty"`

	// Key is the key in the secret containing the GitHub personal access token.
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-zA-Z0-9][a-zA-Z0-9-_.]*[a-zA-Z0-9_.]$
	// +kubebuilder:validation:Required
	Key string `json:"key,omitempty"`
}

type GitHubRepositoryStatus struct {

	// Conditions represent the latest available observations of the GitHubRepository's state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// DefaultBranch is the default branch of the repository.
	// +kubebuilder:validation:MaxLength=250
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=^[a-zA-Z0-9_.-]+(/[a-zA-Z0-9_.-]+)*$
	DefaultBranch string `json:"defaultBranch,omitempty"`
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
