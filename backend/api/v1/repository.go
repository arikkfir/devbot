package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Repository represents a single source code repository hosted remotely (e.g. on GitHub).
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +condition:commons
// +condition:Authenticated,Unauthenticated:AuthenticationFailed,AuthSecretForbidden,AuthSecretKeyNotFound,AuthSecretNotFound,AuthTokenEmpty,InternalError,Invalid
// +condition:Current,Stale:BranchesOutOfSync,DefaultBranchOutOfSync,InternalError,Invalid,RepositoryNotFound,Unauthenticated
// +condition:Valid,Invalid:AuthConfigMissing,AuthSecretKeyMissing,AuthSecretNameMissing,InvalidRefreshInterval,RepositoryNameMissing,RepositoryOwnerMissing,UnknownRepositoryType
// +kubebuilder:printcolumn:name="Refresh Interval",type=string,JSONPath=`.spec.refreshInterval`
// +kubebuilder:printcolumn:name="Valid",type=string,JSONPath=`.status.privateArea.Valid`
// +kubebuilder:printcolumn:name="Authenticated",type=string,JSONPath=`.status.privateArea.Authenticated`
// +kubebuilder:printcolumn:name="Default Branch",type=string,JSONPath=`.status.defaultBranch`
// +kubebuilder:printcolumn:name="Current",type=string,JSONPath=`.status.privateArea.Current`
type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the desired state of the repository.
	// +kubebuilder:validation:Required
	Spec RepositorySpec `json:"spec"`

	// Status is the observed state of the repository.
	// +kubebuilder:validation:Optional
	Status RepositoryStatus `json:"status,omitempty"`
}

// RepositorySpec provides the specification of the remote source code repository.
type RepositorySpec struct {

	// GitHub is the specification for a GitHub repository. Setting this property will mark this repository as a GitHub
	// repository.
	// +kubebuilder:validation:Optional
	GitHub *GitHubRepositorySpec `json:"github,omitempty"`

	// RefreshInterval is the interval at which to refresh the list of branches in the repository. The value should be
	// specified as a duration string, e.g. "5m" for 5 minutes. The default value is "5m".
	// +kubebuilder:default="5m"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Optional
	RefreshInterval string `json:"refreshInterval,omitempty"`
}

// GitHubRepositorySpec provides the specification for a GitHub repository.
type GitHubRepositorySpec struct {

	// Owner is the GitHub user or organization that owns the repository.
	// +kubebuilder:validation:MaxLength=39
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-zA-Z0-9][a-zA-Z0-9-_]*[a-zA-Z0-9_]$
	// +kubebuilder:validation:Required
	Owner string `json:"owner"`

	// Name is the name of the repository.
	// +kubebuilder:validation:MaxLength=100
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-zA-Z0-9_][a-zA-Z0-9-_.]*[a-zA-Z0-9_.]$
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// PersonalAccessToken signals that we should use a GitHub personal access token (PAT) when accessing the repository
	// and specifies the Kubernetes secret & key that house the token (namespace is optional and will default to the
	// repository's namespace if missing).
	// +kubebuilder:validation:Optional
	PersonalAccessToken *GitHubRepositoryPersonalAccessToken `json:"personalAccessToken,omitempty"`
}

// GitHubRepositoryPersonalAccessToken specifies the Kubernetes secret & key that house the GitHub personal access token
// to be used to access the repository.
type GitHubRepositoryPersonalAccessToken struct {

	// Secret is the reference to the secret containing the GitHub personal access token.
	// +kubebuilder:validation:Required
	Secret SecretReferenceWithOptionalNamespace `json:"secret"`

	// Key is the key in the secret containing the GitHub personal access token.
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-zA-Z0-9][a-zA-Z0-9-_.]*[a-zA-Z0-9_.]$
	// +kubebuilder:validation:Required
	Key string `json:"key"`
}

// RepositoryStatus represents the observed state of the Repository.
type RepositoryStatus struct {

	// Conditions represent the latest available observations of the GitHubRepository's state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// DefaultBranch is the default branch of the repository.
	// +kubebuilder:validation:MaxLength=250
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=^[a-zA-Z0-9_.-]+(/[a-zA-Z0-9_.-]+)*$
	DefaultBranch string `json:"defaultBranch,omitempty"`

	// ResolvedName is a universal human-readable name of the repository. The format of this field can vary depending on
	// the type of repository (e.g. GitHub, GitLab, Bitbucket, etc.).
	// +kubebuilder:validation:Optional
	ResolvedName string `json:"resolvedName,omitempty"`

	// Revisions is a map of branch names to their last detected revision.
	// +kubebuilder:validation:Optional
	Revisions map[string]string `json:"revisions,omitempty"`

	// PrivateArea is not meant for public consumption, nor is it part of the public API. It is exposed due to Go and
	// controller-runtime limitations but is an internal part of the implementation.
	PrivateArea ConditionsInverseState `json:"privateArea,omitempty"`
}

type RepositoryStatusPrivateArea struct {
	Initialized   string `json:"-"`
	Finalized     string `json:"-"`
	Valid         string `json:"valid,omitempty"`
	Authenticated string `json:"authenticated,omitempty"`
	Current       string `json:"current,omitempty"`
}

// RepositoryList contains a list of Repository objects.
// +kubebuilder:object:root=true
type RepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Repository{}, &RepositoryList{})
}
