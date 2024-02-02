package k8s

import (
	"context"
	"github.com/secureworks/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	OwnershipIndexField = "metadata.ownerReference.controller"
)

type ownedBy struct {
	scheme *runtime.Scheme
	owner  client.Object
}

func (o *ownedBy) ApplyToList(opts *client.ListOptions) {
	if o.scheme == nil {
		panic(errors.New("scheme in OwnedBy list option is nil"))
	}

	gvk, err := apiutil.GVKForObject(o.owner, o.scheme)
	if err != nil {
		panic(errors.New("failed to get GVK for object '%T': %w", o, err))
	}

	ownerName := o.owner.GetName()
	ownerNamespace := o.owner.GetNamespace()
	ownerGVKAndName := gvk.GroupVersion().String() + "/" + gvk.Kind + ":" + ownerNamespace + "/" + ownerName

	mf := client.MatchingFields{OwnershipIndexField: ownerGVKAndName}
	mf.ApplyToList(opts)
}

func OwnedBy(scheme *runtime.Scheme, owner client.Object) client.ListOption {
	return &ownedBy{owner: owner, scheme: scheme}
}

func AddOwnershipIndex(ctx context.Context, mgr manager.Manager, objType client.Object) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, objType, OwnershipIndexField, IndexGetOwnerReferencesOf); err != nil {
		return errors.New("failed to index '%s' of '%T'", OwnershipIndexField, objType, err)
	}
	return nil
}

func IndexGetOwnerReferencesOf(obj client.Object) []string {
	var owners []string
	for _, ownerReference := range obj.GetOwnerReferences() {
		gvk := schema.FromAPIVersionAndKind(ownerReference.APIVersion, ownerReference.Kind)
		owner := gvk.GroupVersion().String() + "/" + gvk.Kind + ":" + obj.GetNamespace() + "/" + ownerReference.Name
		owners = append(owners, owner)
	}
	return owners
}
