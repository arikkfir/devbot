package reconcile

import (
	"context"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"github.com/secureworks/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewGetOwnedObjectsOfTypeAction(owned client.ObjectList) Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		var status k8s.CommonStatus
		if statusProvider, ok := o.(k8s.CommonStatusProvider); ok {
			status = statusProvider.GetStatus()
		} else {
			panic(errors.New("type '%T' does not implement k8s.CommonStatusProvider", o))
		}

		if err := c.List(ctx, owned, k8s.OwnedBy(c.Scheme(), o)); err != nil {
			status.SetInvalidDueToFailedGettingOwnedObjects("Failed listing owned objects: %+v", err)
			return RequeueDueToError(errors.New("failed listing owned objects: %w", err))
		}

		return Continue()
	}
}
