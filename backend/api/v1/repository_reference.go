package v1

import (
	"github.com/secureworks/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RepositoryReferenceWithOptionalNamespace struct {
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	APIVersion string `json:"apiVersion,omitempty"`

	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Kind string `json:"kind,omitempty"`

	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-z0-9]+(\-[a-z0-9]+)*$
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=^[a-z0-9]+(\-[a-z0-9]+)*$
	Namespace string `json:"namespace,omitempty"`
}

func (in *RepositoryReferenceWithOptionalNamespace) GetObjectKey(defaultNamespace string) client.ObjectKey {
	namespace := in.Namespace
	if namespace == "" {
		namespace = defaultNamespace
	}
	return client.ObjectKey{Namespace: namespace, Name: in.Name}
}

func (in *RepositoryReferenceWithOptionalNamespace) GetGVK() (v1.GroupVersionKind, error) {
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

func (in *RepositoryReferenceWithOptionalNamespace) IsGitHubRepository() bool {
	return in.APIVersion == GitHubRepositoryGVK.GroupVersion().String() && in.Kind == GitHubRepositoryGVK.Kind
}

func (in *RepositoryReferenceWithOptionalNamespace) ToNamespacedRepositoryReference(defaultNamespace string) NamespacedRepositoryReference {
	namespace := in.Namespace
	if namespace == "" {
		namespace = defaultNamespace
	}
	return NamespacedRepositoryReference{
		APIVersion: in.APIVersion,
		Kind:       in.Kind,
		Name:       in.Name,
		Namespace:  namespace,
	}
}

type NamespacedRepositoryReference struct {
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	APIVersion string `json:"apiVersion,omitempty"`

	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Kind string `json:"kind,omitempty"`

	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-z0-9]+(\-[a-z0-9]+)*$
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=^[a-z0-9]+(\-[a-z0-9]+)*$
	Namespace string `json:"namespace,omitempty"`
}

func (in *NamespacedRepositoryReference) GetObjectKey() client.ObjectKey {
	return client.ObjectKey{Namespace: in.Namespace, Name: in.Name}
}

func (in *NamespacedRepositoryReference) GetGVK() (v1.GroupVersionKind, error) {
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

func (in *NamespacedRepositoryReference) IsGitHubRepository() bool {
	return in.APIVersion == GitHubRepositoryGVK.GroupVersion().String() && in.Kind == GitHubRepositoryGVK.Kind
}
