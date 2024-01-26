package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ObjectWithoutCommonConditions is a generic object with conditions.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +condition:Valid,Invalid:C1,C2
type ObjectWithoutCommonConditions struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Optional
	Spec ObjectWithoutCommonConditionsSpec `json:"spec,omitempty"`

	// +kubebuilder:validation:Optional
	Status ObjectWithoutCommonConditionsStatus `json:"status,omitempty"`
}

type ObjectWithoutCommonConditionsSpec struct{}

type ObjectWithoutCommonConditionsStatus struct {
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true

type ObjectWithoutCommonConditionsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ObjectWithoutCommonConditions `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ObjectWithoutCommonConditions{}, &ObjectWithoutCommonConditionsList{})
}
