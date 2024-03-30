package v1

import "sigs.k8s.io/controller-runtime/pkg/client"

type SecretReferenceWithOptionalNamespace struct {
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
}

func (in *SecretReferenceWithOptionalNamespace) GetObjectKey(defaultNamespace string) client.ObjectKey {
	namespace := in.Namespace
	if namespace == "" {
		namespace = defaultNamespace
	}
	return client.ObjectKey{Namespace: namespace, Name: in.Name}
}

func (in *ApplicationSpecRepository) GetObjectKey(defaultNamespace string) client.ObjectKey {
	namespace := in.Namespace
	if namespace == "" {
		namespace = defaultNamespace
	}
	return client.ObjectKey{Namespace: namespace, Name: in.Name}
}
