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

func NewAddFinalizerAction(finalizer string) Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		var status k8s.CommonStatus
		if statusProvider, ok := o.(k8s.CommonStatusProvider); ok {
			status = statusProvider.GetStatus()
		} else {
			panic(errors.New("type '%T' does not implement k8s.CommonStatusProvider", o))
		}

		if !slices.Contains(o.GetFinalizers(), finalizer) {
			o.SetFinalizers(append(o.GetFinalizers(), finalizer))
			if err := c.Update(ctx, o); err != nil {
				if apierrors.IsNotFound(err) {
					return DoNotRequeue()
				} else if apierrors.IsConflict(err) || apierrors.IsGone(err) {
					return Requeue()
				} else {
					status.SetInvalidDueToAddFinalizerFailed("Failed adding finalizer: %+v", err)
					return Requeue()
				}
			}
			return Requeue()
		}

		return Continue()
	}
}
