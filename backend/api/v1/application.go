package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +StatusCondition:Application:Valid

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

type ApplicationSpec struct {
	// Repositories point to the set of
	Repositories []corev1.ObjectReference `json:"repositories,omitempty"`
	Strategy     ApplicationSpecStrategy  `json:"strategy,omitempty"`
}

type ApplicationSpecStrategy struct {
	//+kubebuilder:validation:Enum=Ignore;Required;Default
	Missing       string `json:"missing,omitempty"`
	DefaultBranch string `json:"defaultBranch,omitempty"`
}

type ApplicationStatus struct {
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
