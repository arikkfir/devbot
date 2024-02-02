package reconciler

import (
	"context"
	v1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/secureworks/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"slices"
	"time"
)

type FinalizingObjectStatus interface {
	IsFinalizing() bool
	SetFinalizingDueToFinalizationFailed(message string, args ...interface{}) bool
	SetFinalizingDueToFinalizerRemovalFailed(message string, args ...interface{}) bool
	SetFinalizingDueToInProgress(message string, args ...interface{}) bool
	SetFinalizedIfFinalizingDueToAnyOf(reason ...string) bool
}

type InitializableObjectStatus interface {
	SetFailedToInitializeDueToInternalError(message string, args ...interface{}) bool
	SetInitializedIfFailedToInitializeDueToAnyOf(reason ...string) bool
}

type ControlleeObjectStatus interface {
	SetInvalidDueToControllerNotAccessible(message string, args ...interface{}) bool
	SetInvalidDueToControllerNotFound(message string, args ...interface{}) bool
	SetInvalidDueToControllerReferenceMissing(message string, args ...interface{}) bool
	SetInvalidDueToInternalError(message string, args ...interface{}) bool
	SetValidIfInvalidDueToAnyOf(reason ...string) bool
}

type Reconciliation[O client.Object] struct {
	Ctx            context.Context
	Client         client.Client
	Object         O
	FinalizerValue string
	FinalizerFunc  func(context.Context, client.Client, O) error
}

func NewReconciliation[O client.Object](ctx context.Context, c client.Client, req ctrl.Request, object O, finalizerValue string, finalizerFunc func(context.Context, client.Client, O) error) (*Reconciliation[O], *Result) {
	if err := c.Get(ctx, req.NamespacedName, object); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, DoNotRequeue()
		} else {
			return nil, RequeueDueToError(errors.New("failed to get '%s'", req.NamespacedName, err))
		}
	}
	return &Reconciliation[O]{
		Ctx:            ctx,
		Client:         c,
		Object:         object,
		FinalizerValue: finalizerValue,
		FinalizerFunc:  finalizerFunc,
	}, nil
}

func (r *Reconciliation[O]) FinalizeObjectIfDeleted() *Result {
	status := MustGetStatusOfType[FinalizingObjectStatus](r.Object)

	if r.Object.GetDeletionTimestamp() != nil {
		if slices.Contains(r.Object.GetFinalizers(), r.FinalizerValue) {

			status.SetFinalizingDueToInProgress("Finalizing")
			if result := r.UpdateStatus(WithStrategy(Continue)); result != nil {
				return result
			}

			if r.FinalizerFunc != nil {
				if err := r.FinalizerFunc(r.Ctx, r.Client, r.Object); err != nil {
					status.SetFinalizingDueToFinalizationFailed("%+v", err)
					return r.UpdateStatus(WithStrategy(Requeue))
				}
			}

			r.Object.SetFinalizers(slices.DeleteFunc(r.Object.GetFinalizers(), func(s string) bool { return s == r.FinalizerValue }))

			if err := r.Client.Update(r.Ctx, r.Object); err != nil {
				if apierrors.IsNotFound(err) {
					return DoNotRequeue()
				} else if apierrors.IsConflict(err) || apierrors.IsGone(err) {
					return Requeue()
				} else {
					status.SetFinalizingDueToFinalizerRemovalFailed("%+v", err)
					return r.UpdateStatus(WithStrategy(Requeue))
				}
			}
		}

		status.SetFinalizedIfFinalizingDueToAnyOf(v1.InProgress, v1.FinalizationFailed, v1.FinalizerRemovalFailed)
		return r.UpdateStatus(WithStrategy(DoNotRequeue))
	}

	return Continue()
}

func (r *Reconciliation[O]) InitializeObject() *Result {
	status := MustGetStatusOfType[InitializableObjectStatus](r.Object)

	if !slices.Contains(r.Object.GetFinalizers(), r.FinalizerValue) {

		r.Object.SetFinalizers(append(r.Object.GetFinalizers(), r.FinalizerValue))

		if err := r.Client.Update(r.Ctx, r.Object); err != nil {
			if apierrors.IsNotFound(err) {
				return DoNotRequeue()
			} else if apierrors.IsConflict(err) || apierrors.IsGone(err) {
				return Requeue()
			} else {
				status.SetFailedToInitializeDueToInternalError("Failed adding finalizer: %+v", err)
				return r.UpdateStatus(WithStrategy(Requeue))
			}
		}
	}

	status.SetInitializedIfFailedToInitializeDueToAnyOf(v1.InternalError)
	return r.UpdateStatus(WithStrategy(Continue))
}

func (r *Reconciliation[O]) UpdateStatus(strategy *ErrorStrategy) *Result {
	// TODO: find a way to prevent excessive status updates; this is called in every action in every reconciliation, multiple times

	// Ensure that newly-set conditions have the correct ObservedGeneration and LastTransitionTime
	if conditionsProvider, ok := GetStatusOfType[ConditionsProvider](r.Object); ok {
		var newConditions []metav1.Condition
		for _, c := range conditionsProvider.GetConditions() {
			nc := c
			if nc.ObservedGeneration == 0 {
				nc.ObservedGeneration = r.Object.GetGeneration()
			}
			if nc.LastTransitionTime.IsZero() {
				nc.LastTransitionTime = metav1.Now()
			}
			newConditions = append(newConditions, nc)
		}
		conditionsProvider.SetConditions(newConditions)
	}

	// Update the status subresource, invoking the error strategy accordingly if an error occurs
	if err := r.Client.Status().Update(r.Ctx, r.Object); err != nil {
		if apierrors.IsNotFound(err) {
			return strategy.OnSuccess()
		} else if apierrors.IsConflict(err) || apierrors.IsGone(err) {
			return strategy.OnConflict()
		} else {
			return strategy.OnUnexpectedError(errors.New("failed to update object status: %w", err))
		}
	}
	return strategy.OnSuccess()
}

func (r *Reconciliation[O]) GetRequiredController(controller client.Object) *Result {
	status := MustGetStatusOfType[ControlleeObjectStatus](r.Object)

	controllerRef := metav1.GetControllerOf(r.Object)
	if controllerRef == nil {
		status.SetInvalidDueToControllerReferenceMissing("Controller reference not found")
		return r.UpdateStatus(WithStrategy(DoNotRequeue))
	}

	controllerKey := client.ObjectKey{Name: controllerRef.Name, Namespace: r.Object.GetNamespace()}
	if err := r.Client.Get(r.Ctx, controllerKey, controller); err != nil {
		if apierrors.IsNotFound(err) {
			status.SetInvalidDueToControllerNotFound("Controller object not found")
			return r.UpdateStatus(WithStrategy(Requeue))
		} else if apierrors.IsForbidden(err) {
			status.SetInvalidDueToControllerNotAccessible("Controller object not accessible: %+v", err)
			return r.UpdateStatus(WithStrategy(Requeue))
		} else {
			status.SetInvalidDueToInternalError("Failed getting controller: %+v", err)
			return r.UpdateStatus(WithStrategy(Requeue))
		}
	}

	if status.SetValidIfInvalidDueToAnyOf(v1.ControllerReferenceMissing, v1.ControllerNotFound, v1.ControllerNotAccessible, v1.InternalError) {
		return r.UpdateStatus(WithStrategy(Continue))
	} else {
		return Continue()
	}
}

type RefreshIntervalParsingStatus interface {
	GetInvalidMessage() string
	SetInvalidDueToInvalidRefreshInterval(string, ...interface{}) bool
	SetValidIfInvalidDueToAnyOf(...string) bool
	SetValidIfInvalidDueToInvalidRefreshInterval() bool
}

func (r *Reconciliation[O]) ParseRefreshInterval(value string, targetRefreshInterval *time.Duration, statusUpdateCallbacks ...func(string, ...interface{}) bool) *Result {
	if duration, err := time.ParseDuration(value); err != nil {
		for _, callback := range statusUpdateCallbacks {
			callback("Refresh interval is invalid: %+v", err)
		}
		return r.UpdateStatus(WithStrategy(DoNotRequeue))
	} else if duration.Seconds() < 5 {
		for _, callback := range statusUpdateCallbacks {
			callback("Refresh interval '%s' is too low (must not be less than 5 seconds)", value)
		}
		return r.UpdateStatus(WithStrategy(DoNotRequeue))
	} else {
		*targetRefreshInterval = duration
		return Continue()
	}
}
