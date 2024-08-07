package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Environment is the Schema for the environments API.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +condition:commons
// +condition:Current,Stale:DeploymentsAreStale,FailedCreatingDeployment,FailedDeletingDeployment,InternalError
// +kubebuilder:printcolumn:name="Preferred Branch",type=string,JSONPath=`.spec.branch`
// +kubebuilder:printcolumn:name="Valid",type=string,JSONPath=`.status.privateArea.Valid`
// +kubebuilder:printcolumn:name="Current",type=string,JSONPath=`.status.privateArea.Current`
type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the desired state of the Environment.
	// +kubebuilder:validation:Required
	Spec EnvironmentSpec `json:"spec"`

	// Status is the observed state of the Environment.
	// +kubebuilder:validation:Optional
	Status EnvironmentStatus `json:"status,omitempty"`
}

type EnvironmentSpec struct {
	// PreferredBranch is the preferred branch for deployment to this environment from each repository. Repositories
	// that lack this branch may opt to deploy their default branch instead (see [ApplicationSpecRepository.MissingBranchStrategy]).
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	PreferredBranch string `json:"branch"`
}

type EnvironmentStatus struct {

	// Conditions represent the latest available observations of the application environment's state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// PrivateArea is not meant for public consumption, nor is it part of the public API. It is exposed due to Go and
	// controller-runtime limitations but is an internal part of the implementation.
	PrivateArea ConditionsInverseState `json:"privateArea,omitempty"`
}

// +kubebuilder:object:root=true

type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Environment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Environment{}, &EnvironmentList{})
}
