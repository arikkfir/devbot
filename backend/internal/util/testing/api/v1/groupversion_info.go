// Package v1 contains API Schema definitions for the testing.devbot.kfirs.com v1 API group
// +kubebuilder:object:generate=true
// +groupName=testing.devbot.kfirs.com
//
//go:generate controller-gen object paths="."
//go:generate go run ../../../../../scripts/generators/conditions .
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	Domain  = "testing.devbot.kfirs.com"
	Version = "v1"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: Domain, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

var (
	ObjectWithCommonConditionsGVK    = schema.GroupVersionKind{Group: Domain, Version: Version, Kind: "ObjectWithCommonConditions"}
	ObjectWithoutCommonConditionsGVK = schema.GroupVersionKind{Group: Domain, Version: Version, Kind: "ObjectWithoutCommonConditions"}
)
