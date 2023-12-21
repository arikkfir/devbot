// Package v1 contains API Schema definitions for the devbot.kfirs.com v1 API group
// +kubebuilder:object:generate=true
// +groupName=devbot.kfirs.com
//
//go:generate controller-gen object crd paths="." output:crd:artifacts:config=../../../deploy/app/crd
//go:generate go run ../../scripts/generators/api-status-conditions/main.go
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	ConditionTypeAuthenticatedToGitHub = "AuthenticatedToGitHub"
	ConditionTypeCurrent               = "Current"
	ConditionTypeDeploying             = "Deploying"
	ConditionTypeValid                 = "Valid"
	ReasonAuthenticated                = "Authenticated"
	ReasonConfigError                  = "ConfigError"
	ReasonAuthConfigError              = "AuthConfigError"
	ReasonGitHubAuthSecretNameMissing  = "GitHubAuthSecretNameMissing"
	ReasonGitHubAuthSecretNotFound     = "GitHubAuthSecretNotFound"
	ReasonGitHubAuthSecretForbidden    = "GitHubAuthSecretForbidden"
	ReasonGitHubAuthSecretEmptyToken   = "GitHubAuthSecretEmptyToken"
	ReasonGitHubAPIFailed              = "GitHubAPIFailed"
	ReasonInitializing                 = "Initializing"
	ReasonInternalError                = "InternalError"
	ReasonSynced                       = "Synced"
	ReasonValid                        = "Valid"
)

const (
	Domain  = "devbot.kfirs.com"
	Version = "v1"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: Domain, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

var (
	ApplicationGVK         = schema.GroupVersionKind{Group: Domain, Version: Version, Kind: "Application"}
	GitHubRepositoryGVK    = schema.GroupVersionKind{Group: Domain, Version: Version, Kind: "GitHubRepository"}
	GitHubRepositoryRefGVK = schema.GroupVersionKind{Group: Domain, Version: Version, Kind: "GitHubRepositoryRef"}
)
