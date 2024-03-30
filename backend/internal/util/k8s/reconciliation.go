package k8s

import (
	"context"
	v1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/config"
	"github.com/secureworks/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"slices"
)

type ConditionsProvider interface {
	GetConditions() []metav1.Condition
	SetConditions(conditions []metav1.Condition)
}

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
	Config         config.CommandConfig
	Client         client.Client
	Object         O
	finalizerValue string
	finalizerFunc  func(context.Context, client.Client, O) error
}

func NewReconciliation[O client.Object](ctx context.Context, cfg config.CommandConfig, c client.Client, req ctrl.Request, object O, finalizerValue string, finalizerFunc func(context.Context, client.Client, O) error) (*Reconciliation[O], *Result) {
	if err := c.Get(ctx, req.NamespacedName, object); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, DoNotRequeue()
		} else {
			return nil, RequeueDueToError(errors.New("failed to get '%s'", req.NamespacedName, err))
		}
	}
	return &Reconciliation[O]{
		Ctx:            ctx,
		Config:         cfg,
		Client:         c,
		Object:         object,
		finalizerValue: finalizerValue,
		finalizerFunc:  finalizerFunc,
	}, nil
}

func (r *Reconciliation[O]) UpdateStatus() *Result {
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
			return DoNotRequeue()
		} else if apierrors.IsConflict(err) {
			return Requeue()
		} else {
			return RequeueDueToError(errors.New("failed to update object status: %w", err))
		}
	}
	return nil
}

func (r *Reconciliation[O]) FinalizeObjectIfDeleted() *Result {
	status := MustGetStatusOfType[FinalizingObjectStatus](r.Object)

	if r.Object.GetDeletionTimestamp() != nil {
		if slices.Contains(r.Object.GetFinalizers(), r.finalizerValue) {

			status.SetFinalizingDueToInProgress("Finalizing")
			if result := r.UpdateStatus(); result != nil {
				return result
			}

			if r.finalizerFunc != nil {
				if err := r.finalizerFunc(r.Ctx, r.Client, r.Object); err != nil {
					status.SetFinalizingDueToFinalizationFailed("%+v", err)
					if result := r.UpdateStatus(); result != nil {
						return result
					}
					return Requeue()
				}
			}

			r.Object.SetFinalizers(slices.DeleteFunc(r.Object.GetFinalizers(), func(s string) bool { return s == r.finalizerValue }))

			if err := r.Client.Update(r.Ctx, r.Object); err != nil {
				if apierrors.IsNotFound(err) {
					return DoNotRequeue()
				} else if apierrors.IsConflict(err) || apierrors.IsGone(err) {
					return Requeue()
				} else {
					status.SetFinalizingDueToFinalizerRemovalFailed("%+v", err)
					if result := r.UpdateStatus(); result != nil {
						return result
					}
					return Requeue()
				}
			}
		}

		status.SetFinalizedIfFinalizingDueToAnyOf(v1.InProgress, v1.FinalizationFailed, v1.FinalizerRemovalFailed)
		if result := r.UpdateStatus(); result != nil {
			return result
		}
		return DoNotRequeue()
	}

	return Continue()
}

func (r *Reconciliation[O]) InitializeObject() *Result {
	status := MustGetStatusOfType[InitializableObjectStatus](r.Object)

	if !slices.Contains(r.Object.GetFinalizers(), r.finalizerValue) {

		r.Object.SetFinalizers(append(r.Object.GetFinalizers(), r.finalizerValue))

		if err := r.Client.Update(r.Ctx, r.Object); err != nil {
			if apierrors.IsNotFound(err) {
				return DoNotRequeue()
			} else if apierrors.IsConflict(err) || apierrors.IsGone(err) {
				return Requeue()
			} else {
				status.SetFailedToInitializeDueToInternalError("Failed adding finalizer: %+v", err)
				if result := r.UpdateStatus(); result != nil {
					return result
				}
				return Requeue()
			}
		}
	}

	status.SetInitializedIfFailedToInitializeDueToAnyOf(v1.InternalError)
	if result := r.UpdateStatus(); result != nil {
		return result
	}
	return Continue()
}

func (r *Reconciliation[O]) GetRequiredController(controller client.Object) *Result {
	status := MustGetStatusOfType[ControlleeObjectStatus](r.Object)

	controllerRef := metav1.GetControllerOf(r.Object)
	if controllerRef == nil {
		status.SetInvalidDueToControllerReferenceMissing("Controller reference not found")
		if result := r.UpdateStatus(); result != nil {
			return result
		}
		return DoNotRequeue()
	}

	controllerKey := client.ObjectKey{Name: controllerRef.Name, Namespace: r.Object.GetNamespace()}
	if err := r.Client.Get(r.Ctx, controllerKey, controller); err != nil {
		if apierrors.IsNotFound(err) {
			status.SetInvalidDueToControllerNotFound("Controller object not found")
			if result := r.UpdateStatus(); result != nil {
				return result
			}
			return Requeue()
		} else if apierrors.IsForbidden(err) {
			status.SetInvalidDueToControllerNotAccessible("Controller object not accessible: %+v", err)
			if result := r.UpdateStatus(); result != nil {
				return result
			}
			return Requeue()
		} else {
			status.SetInvalidDueToInternalError("Failed getting controller: %+v", err)
			if result := r.UpdateStatus(); result != nil {
				return result
			}
			return Requeue()
		}
	}

	status.SetValidIfInvalidDueToAnyOf(v1.ControllerReferenceMissing, v1.ControllerNotFound, v1.ControllerNotAccessible, v1.InternalError)
	if result := r.UpdateStatus(); result != nil {
		return result
	}
	return Continue()
}
