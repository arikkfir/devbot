package util

import (
	"context"
	"github.com/secureworks/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"slices"
)

const OwnerRefField = "metadata.ownerRef"

func PrepareReconciliation[T client.Object](ctx context.Context, r client.Client, req ctrl.Request, object T, finalizer string) (*ctrl.Result, error) {
	if err := r.Get(ctx, req.NamespacedName, object); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return &ctrl.Result{}, nil
		} else {
			return &ctrl.Result{}, errors.New("failed to get '%s'", req.NamespacedName, err)
		}
	}

	if object.GetDeletionTimestamp() != nil {
		if slices.Contains(object.GetFinalizers(), finalizer) {
			var finalizers []string
			for _, f := range object.GetFinalizers() {
				if f != finalizer {
					finalizers = append(finalizers, f)
				}
			}
			object.SetFinalizers(finalizers)
			if err := r.Update(ctx, object); err != nil {
				return &ctrl.Result{}, err
			}
		}
		return &ctrl.Result{}, nil
	}

	// Add finalizer for this CR if it's missing
	if !slices.Contains(object.GetFinalizers(), finalizer) {
		object.SetFinalizers(append(object.GetFinalizers(), finalizer))
		if err := r.Update(ctx, object); err != nil {
			return &ctrl.Result{}, err
		}
		return &ctrl.Result{Requeue: true}, nil
	}

	// Ready - return nil, nil to signal that processing should simply continue
	return nil, nil
}

func GetOwnerRefKey(o client.Object) string {
	gvk := o.GetObjectKind().GroupVersionKind()
	gvkAndName := gvk.GroupVersion().String() + "/" + gvk.Kind + ":" + o.GetNamespace() + "/" + o.GetName()
	return gvkAndName
}

func IndexableOwnerReferences(obj client.Object) []string {
	var owners []string
	for _, ownerReference := range obj.GetOwnerReferences() {
		gvk := schema.FromAPIVersionAndKind(ownerReference.APIVersion, ownerReference.Kind)
		owner := gvk.GroupVersion().String() + "/" + gvk.Kind + ":" + obj.GetNamespace() + "/" + ownerReference.Name
		owners = append(owners, owner)
	}
	return owners
}

func GetOwnerOfObject(ctx context.Context, c client.Client, obj client.Object, expectedAPIVersion, expectedKind string, owner client.Object) error {
	var ownerRef *metav1.OwnerReference
	for _, or := range obj.GetOwnerReferences() {
		if or.APIVersion == expectedAPIVersion && or.Kind == expectedKind {
			ownerRef = &or
			break
		}
	}

	if ownerRef == nil {
		return errors.New("could not find owner of kind '%s.%s'", expectedKind, expectedAPIVersion)
	}

	if err := c.Get(ctx, client.ObjectKey{Namespace: obj.GetNamespace(), Name: ownerRef.Name}, owner); err != nil {
		return errors.New("failed to get owner '%s/%s': %w", obj.GetNamespace(), ownerRef.Name, err)
	}

	return nil
}
