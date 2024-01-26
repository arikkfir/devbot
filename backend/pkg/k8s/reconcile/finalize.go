package reconcile

import (
	"context"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"github.com/secureworks/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"slices"
)

func NewFinalizeAction(finalizer string, finalize func() error) Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		var status k8s.CommonStatus
		if statusProvider, ok := o.(k8s.CommonStatusProvider); ok {
			status = statusProvider.GetStatus()
		} else {
			panic(errors.New("type '%T' does not implement k8s.CommonStatusProvider", o))
		}

		if o.GetDeletionTimestamp() != nil {
			if slices.Contains(o.GetFinalizers(), finalizer) {
				if finalize != nil {
					if err := finalize(); err != nil {
						status.SetInvalidDueToFinalizationFailed("Finalization failed: %+v", err)
						return Requeue()
					}
				}

				o.SetFinalizers(slices.DeleteFunc(o.GetFinalizers(), func(s string) bool { return s == finalizer }))

				if err := c.Update(ctx, o); err != nil {
					if apierrors.IsNotFound(err) {
						return DoNotRequeue()
					} else if apierrors.IsConflict(err) || apierrors.IsGone(err) {
						return Requeue()
					} else {
						status.SetInvalidDueToFinalizationFailed("Failed removing finalizer: %+v", err)
						return Requeue()
					}
				}
			}
			return DoNotRequeue()
		}

		return Continue()
	}
}
