package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ObjectWithCommonConditions is a generic object with conditions.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +condition:Controlled,Uncontrolled:ControllerNotAccessible,ControllerNotFound,ControllerCannotBeFetched,OwnerReferenceMissing
// +condition:Finalized,Finalizing:FinalizationFailed,FinalizerRemovalFailed,InProgress
// +condition:Initialized,FailedToInitialize:InternalError
type ObjectWithCommonConditions struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Optional
	Spec ObjectWithCommonConditionsSpec `json:"spec,omitempty"`

	// +kubebuilder:validation:Optional
	Status ObjectWithCommonConditionsStatus `json:"status,omitempty"`
}

type ObjectWithCommonConditionsSpec struct{}

type ObjectWithCommonConditionsStatus struct {
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true

type ObjectWithCommonConditionsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ObjectWithCommonConditions `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ObjectWithCommonConditions{}, &ObjectWithCommonConditionsList{})
}
