package v1

import "sigs.k8s.io/controller-runtime/pkg/client"

type DeploymentRepositoryReference struct {
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-z0-9]+(\-[a-z0-9]+)*$
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=^[a-z0-9]+(\-[a-z0-9]+)*$
	Namespace string `json:"namespace,omitempty"`
}

func (in *DeploymentRepositoryReference) GetObjectKey() client.ObjectKey {
	return client.ObjectKey{Namespace: in.Namespace, Name: in.Name}
}
