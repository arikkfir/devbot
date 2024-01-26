package reconcile

import (
	"context"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/secureworks/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

type Reconciliation struct {
	Actions []Action
}

func (rec *Reconciliation) Execute(ctx context.Context, c client.Client, req ctrl.Request, o client.Object) (ctrl.Result, error) {
	if err := c.Get(ctx, req.NamespacedName, o); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, errors.New("failed to get '%s'", req.NamespacedName, err)
		}
	}

	for _, action := range rec.Actions {
		result, err := action.Execute(ctx, c, o)
		if result != nil && err != nil {
			return *result, err
		} else if err != nil {
			return ctrl.Result{}, err
		} else if result != nil {
			return *result, nil
		}
	}

	return ctrl.Result{}, nil
}

type Action func(context.Context, client.Client, client.Object) (*ctrl.Result, error)

func (a Action) Execute(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
	statusProvider, ok := o.(k8s.CommonStatusProvider)
	if !ok {
		// No access to conditions; just execute the action and return its result
		return a(ctx, c, o)
	}
	oldObject := statusProvider.DeepCopyObject().(k8s.CommonStatusProvider)
	statusProvider.GetStatus().SetConditions(nil)

	// Execute the action
	result, err := a(ctx, c, o)

	// Ensure that newly-set conditions have the correct ObservedGeneration and LastTransitionTime
	newStatus := statusProvider.GetStatus()
	var newConditions []metav1.Condition
	for _, c := range newStatus.GetConditions() {
		nc := c
		if nc.ObservedGeneration == 0 {
			nc.ObservedGeneration = o.GetGeneration()
		}
		if nc.LastTransitionTime.IsZero() {
			for _, oc := range oldObject.GetStatus().GetConditions() {
				if oc.Type == nc.Type {
					nc.LastTransitionTime = oc.LastTransitionTime
					break
				}
			}
		}
		if nc.LastTransitionTime.IsZero() {
			nc.LastTransitionTime = metav1.Now()
		}
		newConditions = append(newConditions, nc)
	}
	newStatus.SetConditions(newConditions)

	// Compare old & new status; update status if it has changed
	if !cmp.Equal(oldObject.GetStatus(), newStatus, cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime", "ObservedGeneration")) {
		if statusErr := c.Status().Update(ctx, o); statusErr != nil {
			if apierrors.IsNotFound(statusErr) {
				return DoNotRequeue()
			} else if apierrors.IsConflict(statusErr) || apierrors.IsGone(statusErr) {
				return Requeue()
			} else {
				log.FromContext(ctx).Error(err, "Reconciliation failed but overshadowed by status update error; underlying error is logged here, and reconciliation will be retried (subsequent controller message will show the status update error)")
				return RequeueDueToError(errors.New("failed to update object status: %w", statusErr))
			}
		}
	}

	// Return the result and error from the action
	return result, err
}

func Continue() (*ctrl.Result, error) {
	return nil, nil
}

func DoNotRequeue() (*ctrl.Result, error) {
	return &ctrl.Result{}, nil
}

func Requeue() (*ctrl.Result, error) {
	return &ctrl.Result{Requeue: true}, nil
}

func RequeueAfter(interval time.Duration) (*ctrl.Result, error) {
	return &ctrl.Result{RequeueAfter: interval}, nil
}

func RequeueDueToError(err error) (*ctrl.Result, error) {
	return &ctrl.Result{}, err
}
