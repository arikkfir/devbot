// Package v1 contains API Schema definitions for the devbot.kfirs.com v1 API group
// +kubebuilder:object:generate=true
// +groupName=devbot.kfirs.com
//
//go:generate controller-gen object crd paths="." output:crd:artifacts:config=../../../deploy/app/crd
//go:generate go run ../../scripts/generators/conditions .
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
	"time"
)

const (
	Domain                       = "devbot.kfirs.com"
	Version                      = "v1"
	MinRepositoryRefreshInterval = 5 * time.Second
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
	ApplicationGVK = schema.GroupVersionKind{Group: Domain, Version: Version, Kind: "Application"}
	EnvironmentGVK = schema.GroupVersionKind{Group: Domain, Version: Version, Kind: "Environment"}
	RepositoryGVK  = schema.GroupVersionKind{Group: Domain, Version: Version, Kind: "Repository"}
)
