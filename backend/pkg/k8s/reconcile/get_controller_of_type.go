package reconcile

import (
	"context"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"github.com/secureworks/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewGetControllerAction(optional bool, controllee client.Object, controller client.Object) Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		var status k8s.CommonStatus
		if statusProvider, ok := o.(k8s.CommonStatusProvider); ok {
			status = statusProvider.GetStatus()
		} else {
			panic(errors.New("type '%T' does not implement k8s.CommonStatusProvider", o))
		}

		controllerRef := metav1.GetControllerOf(controllee)
		if controllerRef == nil {
			if optional {
				return Continue()
			} else {
				status.SetInvalidDueToControllerMissing("Controller reference not found")
				return DoNotRequeue()
			}
		}

		if err := c.Get(ctx, client.ObjectKey{Name: controllerRef.Name, Namespace: o.GetNamespace()}, controller); err != nil {
			if apierrors.IsNotFound(err) {
				if optional {
					return Continue()
				} else {
					status.SetInvalidDueToControllerMissing("Controller object not found")
					return DoNotRequeue()
				}
			} else {
				status.SetInvalidDueToInternalError("Failed obtaining controller '%s': %+v", controllerRef.Name, err)
				return RequeueDueToError(errors.New("failed to get owner: %w", err))
			}
		}

		return Continue()
	}
}
