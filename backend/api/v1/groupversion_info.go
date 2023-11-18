// Package v1 contains API Schema definitions for the devbot.kfirs.com v1 API group
// +kubebuilder:object:generate=true
// +groupName=devbot.kfirs.com
//
//go:generate controller-gen object crd paths="." output:crd:artifacts:config=../../../deploy/app/crd
//go:generate go run ../../scripts/generators/api-status-conditions/main.go
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "devbot.kfirs.com", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
