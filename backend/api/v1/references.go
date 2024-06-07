package v1

import (
	"github.com/secureworks/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OptionallyNamespacedReference struct {
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	APIVersion string `json:"apiVersion"`

	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Kind string `json:"kind"`

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

func (in *OptionallyNamespacedReference) GetObjectKey(defaultNamespace string) client.ObjectKey {
	namespace := in.Namespace
	if namespace == "" {
		namespace = defaultNamespace
	}
	return client.ObjectKey{Namespace: namespace, Name: in.Name}
}

func (in *OptionallyNamespacedReference) GetGVK() (v1.GroupVersionKind, error) {
	version, err := schema.ParseGroupVersion(in.APIVersion)
	if err != nil {
		return v1.GroupVersionKind{}, errors.New("failed to parse API version '%s': %w", in.APIVersion, err)
	}
	gvk := version.WithKind(in.Kind)
	return v1.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}, nil
}

func (in *OptionallyNamespacedReference) ToNamespacedRepositoryReference(defaultNamespace string) NamespacedReference {
	namespace := in.Namespace
	if namespace == "" {
		namespace = defaultNamespace
	}
	return NamespacedReference{
		APIVersion: in.APIVersion,
		Kind:       in.Kind,
		Name:       in.Name,
		Namespace:  namespace,
	}
}

type NamespacedReference struct {
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	APIVersion string `json:"apiVersion"`

	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Kind string `json:"kind"`

	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-z0-9]+(\-[a-z0-9]+)*$
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=^[a-z0-9]+(\-[a-z0-9]+)*$
	Namespace string `json:"namespace"`
}

func (in *NamespacedReference) GetObjectKey() client.ObjectKey {
	return client.ObjectKey{Namespace: in.Namespace, Name: in.Name}
}

func (in *NamespacedReference) GetGVK() (v1.GroupVersionKind, error) {
	version, err := schema.ParseGroupVersion(in.APIVersion)
	if err != nil {
		return v1.GroupVersionKind{}, errors.New("failed to parse API version '%s': %w", in.APIVersion, err)
	}
	gvk := version.WithKind(in.Kind)
	return v1.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}, nil
}
