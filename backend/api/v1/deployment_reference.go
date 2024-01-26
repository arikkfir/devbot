package v1

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeploymentReference struct {
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-z0-9][a-z0-9-]*[a-z0-9]$
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
}

func (in *DeploymentReference) GetObjectKey(namespace string) client.ObjectKey {
	return client.ObjectKey{Namespace: namespace, Name: in.Name}
}
