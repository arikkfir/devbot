package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApplicationGitHubRepository defines the GitHub repository of the application
type ApplicationGitHubRepository struct {
	Owner string `json:"owner,omitempty"`
	Name  string `json:"name,omitempty"`
}

// ApplicationRepository defines the source repository of the application
type ApplicationRepository struct {
	GitHub *ApplicationGitHubRepository `json:"github,omitempty"`
}

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	// Application source repository
	Repository ApplicationRepository `json:"repository,omitempty"`
}

type RefStatus struct {
	LastAppliedCommit     string `json:"lastAppliedCommit,omitempty"`
	LatestAvailableCommit string `json:"latestAvailableCommit,omitempty"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	Refs map[string]*RefStatus `json:"refs,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
