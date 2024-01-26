package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	MissingBranchStrategyUseDefaultBranch = "UseDefaultBranch"
	MissingBranchStrategyIgnore           = "Ignore"
)

// Application is the Schema for the applications API.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +condition:Current,Stale:InternalError,Invalid,RepositoryNotAccessible,RepositoryNotFound
// +condition:Valid,Invalid:AddFinalizerFailed,ControllerMissing,FailedGettingOwnedObjects,FinalizationFailed,InternalError
// +condition:Valid,Invalid:RepositoryNotSupported
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the desired state of the Application.
	// +kubebuilder:validation:Required
	Spec ApplicationSpec `json:"spec,omitempty"`

	// Status is the observed state of the Application.
	// +kubebuilder:validation:Optional
	Status ApplicationStatus `json:"status,omitempty"`
}

type ApplicationSpec struct {

	// Repositories is a list of repositories to be deployed as part of this application.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:Required
	Repositories []ApplicationSpecRepository `json:"repositories,omitempty"`
}

type ApplicationSpecRepository struct {
	RepositoryReferenceWithOptionalNamespace `json:",inline"`

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
