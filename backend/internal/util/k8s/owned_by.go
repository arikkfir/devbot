package k8s

import (
	"context"
	"fmt"

	"github.com/secureworks/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

const (
	OwnershipIndexField = "metadata.ownerReference"
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
	owner := fmt.Sprintf("%s/%s:%s/%s", gvk.GroupVersion(), gvk.Kind, ownerNamespace, ownerName)

	mf := client.MatchingFields{OwnershipIndexField: owner}
	mf.ApplyToList(opts)
}

func OwnedBy(scheme *runtime.Scheme, owner client.Object) client.ListOption {
	return &ownedBy{owner: owner, scheme: scheme}
}

func AddOwnershipIndex(ctx context.Context, fieldIndexes client.FieldIndexer, ownedObjType client.Object) error {
	if err := fieldIndexes.IndexField(ctx, ownedObjType, OwnershipIndexField, IndexGetOwnerReferencesOf); err != nil {
		return errors.New("failed to index '%s' of '%T'", OwnershipIndexField, ownedObjType, err)
	}
	return nil
}

func IndexGetOwnerReferencesOf(obj client.Object) []string {
	var owners []string
	for _, ownerReference := range obj.GetOwnerReferences() {
		gvk := schema.FromAPIVersionAndKind(ownerReference.APIVersion, ownerReference.Kind)
		owner := fmt.Sprintf("%s/%s:%s/%s", gvk.GroupVersion(), gvk.Kind, obj.GetNamespace(), ownerReference.Name)
		owners = append(owners, owner)
	}
	return owners
}
