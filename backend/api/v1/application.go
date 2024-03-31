package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	UseDefaultBranchStrategy = "UseDefaultBranch"
	IgnoreStrategy           = "Ignore"
)

// Application represents a single application, optionally spanning multiple repositories (or a single one) and manages
// multiple deployment environments, as deducted from the different branches in said repositories.
// +kubebuilder:printcolumn:name="Service Account",type=string,JSONPath=`.spec.serviceAccountName`
// +kubebuilder:printcolumn:name="Invalid",type=string,JSONPath=`.status.conditions[?(@.type=="Invalid")].reason`
// +kubebuilder:printcolumn:name="Stale",type=string,JSONPath=`.status.conditions[?(@.type=="Stale")].reason`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +condition:Current,Stale:EnvironmentsAreStale,InternalError,RepositoryNotAccessible,RepositoryNotFound
// +condition:Finalized,Finalizing:FinalizationFailed,FinalizerRemovalFailed,InProgress
// +condition:Initialized,FailedToInitialize:InternalError
// +condition:Valid,Invalid:InvalidBranchSpecification
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the desired state of the Application.
	// +required
	// +kubebuilder:validation:Required
	Spec ApplicationSpec `json:"spec"`

	// Status is the observed state of the Application.
	// +kubebuilder:validation:Optional
	Status ApplicationStatus `json:"status,omitempty"`
}

type ApplicationSpec struct {

	// Repositories is a list of repositories to be deployed as part of this application.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:Required
	Repositories []ApplicationSpecRepository `json:"repositories"`

	// ServiceAccountName is the name of the service account used by the deployment apply job.
	// +required
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-z0-9]+(\-[a-z0-9]+)*$
	// +kubebuilder:validation:Required
	ServiceAccountName string `json:"serviceAccountName"`

	// List of branch regular expressions to track in the participating repositories. Only branches matching one of the
	// expressions listed here will be considered for deployment. If no expressions are provided, all branches will be
	// tracked.
	// +kubebuilder:validation:Optional
	Branches []string `json:"branches,omitempty"`

	// TODO: Add environment expiry support, comprised of a default expiry time, a per-environment override & stickiness
}

type ApplicationSpecRepository struct {
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-z0-9]+(\-[a-z0-9]+)*$
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=^[a-z0-9]+(\-[a-z0-9]+)*$
	Namespace string `json:"namespace,omitempty"`

	// +kubebuilder:default=deploy
	// +kubebuilder:validation:Optional
	Path string `json:"path,omitempty"`

	// MissingBranchStrategy defines what to do when the desired branch of an environment is missing in this repository.
	// If "UseDefaultBranch" is set, then instead of the desired branch the default branch of the repository is used.
	// If "Ignore" is set, then the repository is ignored and not deployed to the application environment.
	// +kubebuilder:default=UseDefaultBranch
	// +kubebuilder:validation:Enum=Ignore;UseDefaultBranch
	MissingBranchStrategy string `json:"missingBranchStrategy,omitempty"`
}

type ApplicationStatus struct {

	// Conditions represent the latest available observations of the application's state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true

type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
