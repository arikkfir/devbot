package k8s

import (
	"context"
	"github.com/secureworks/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"slices"
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
	ownerGVKAndName := gvk.GroupVersion().String() + "/" + gvk.Kind + ":" + ownerNamespace + "/" + ownerName

	mf := client.MatchingFields{OwnershipIndexField: ownerGVKAndName}
	mf.ApplyToList(opts)
}

func OwnedBy(scheme *runtime.Scheme, owner client.Object) client.ListOption {
	return &ownedBy{owner: owner, scheme: scheme}
}

func GetFirstOwnerOfType(ctx context.Context, c client.Client, child client.Object, apiVersion, kind string, owner client.Object) error {
	references := child.GetOwnerReferences()
	for _, ownerRef := range references {
		if ownerRef.APIVersion == apiVersion && ownerRef.Kind == kind {
			return c.Get(ctx, client.ObjectKey{Name: ownerRef.Name, Namespace: child.GetNamespace()}, owner)
		}
	}
	return errors.New("no owner of type '%s.%s' found for '%s/%s'", kind, apiVersion, child.GetNamespace(), child.GetName())
}

func GetOwnedChildrenByIndex[OwnerType client.Object, ListType client.ObjectList](ctx context.Context, c client.Client, owner OwnerType, list ListType) error {
	return c.List(ctx, list, OwnedBy(c.Scheme(), owner))
}

func GetOwnedChildrenManually[OwnerType client.Object, ListType client.ObjectList](ctx context.Context, c client.Client, owner OwnerType, list ListType) error {
	if err := c.List(ctx, list); err != nil {
		return errors.New("failed to list '%T'", list, err)
	}

	ownerKey := GetOwnerReferenceOf(c.Scheme(), owner)

	valueOfList := reflect.ValueOf(list)
	elemValueOfList := valueOfList.Elem()
	itemsFieldValue := elemValueOfList.FieldByName("Items")
	for i := 0; i < itemsFieldValue.Len(); i++ {
		itemValue := itemsFieldValue.Index(i)
		object := itemValue.Addr().Interface().(client.Object)
		if !slices.Contains(IndexGetOwnerReferencesOf(object), ownerKey) {
			pre := itemsFieldValue.Slice(0, i)
			post := itemsFieldValue.Slice(i+1, itemsFieldValue.Len())
			itemsFieldValue.Set(reflect.AppendSlice(pre, post))
			i--
		}
	}
	return nil
}

func AddOwnershipIndex(ctx context.Context, mgr manager.Manager, objType client.Object) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, objType, OwnershipIndexField, IndexGetOwnerReferencesOf); err != nil {
		return errors.New("failed to index '%s' of '%T'", OwnershipIndexField, objType, err)
	}
	return nil
}

func GetOwnerReferenceOf(s *runtime.Scheme, o client.Object) string {
	gvk, err := apiutil.GVKForObject(o, s)
	if err != nil {
		panic(errors.New("failed to get GVK for object '%T': %w", o, err))
	}

	gvkAndName := gvk.GroupVersion().String() + "/" + gvk.Kind + ":" + o.GetNamespace() + "/" + o.GetName()
	return gvkAndName
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

func NewOwnerReference(owner client.Object, gvk schema.GroupVersionKind, controller *bool) *v1.OwnerReference {
	return &v1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		BlockOwnerDeletion: &[]bool{true}[0],
		Controller:         controller,
	}
}
