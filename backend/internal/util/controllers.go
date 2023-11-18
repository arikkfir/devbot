package util

import (
	"context"
	"github.com/secureworks/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"slices"
)

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
