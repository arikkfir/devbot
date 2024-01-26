package application

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	reconcile2 "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var (
	Finalizer = "applications.finalizers." + apiv1.GroupVersion.Group
)

// Reconciler reconciles an Application object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var envList *apiv1.EnvironmentList
	reconciliation := reconcile.Reconciliation{
		Actions: []reconcile.Action{
			reconcile.NewFinalizeAction(Finalizer, nil),
			reconcile.NewAddFinalizerAction(Finalizer),
			reconcile.NewGetOwnedObjectsOfTypeAction(envList),
			NewCreateMissingEnvironmentsAction(envList),
			reconcile.NewRequeueAfterAction(time.Minute),
		}}
	return reconciliation.Execute(ctx, r.Client, req, &apiv1.Application{})
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Application{}).
		Watches(&apiv1.GitHubRepository{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile2.Request {
			ghRepo := obj.(*apiv1.GitHubRepository)
			if ghRepo.APIVersion == "" {
				panic("APIVersion is empty")
			} else if ghRepo.Kind == "" {
				panic("Kind is empty")
			}

			appsList := &apiv1.ApplicationList{}
			if err := r.List(ctx, appsList); err != nil {
				log.FromContext(ctx).Error(err, "Failed to list applications")
				return nil
			}

			var requests []reconcile2.Request
			for _, app := range appsList.Items {
				for _, appRepo := range app.Spec.Repositories {
					if appRepo.APIVersion == ghRepo.APIVersion && appRepo.Kind == ghRepo.Kind {
						if appRepo.Name == ghRepo.Name && appRepo.Namespace == ghRepo.Namespace {
							key := client.ObjectKey{Namespace: app.Namespace, Name: app.Name}
							requests = append(requests, reconcile2.Request{NamespacedName: key})
						}
					}
				}
			}
			return requests
		})).
		Watches(&apiv1.GitHubRepositoryRef{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile2.Request {
			ghRepoRef := obj.(*apiv1.GitHubRepositoryRef)
			controllerRef := metav1.GetControllerOf(ghRepoRef)
			if controllerRef == nil {
				return nil
			}

			appsList := &apiv1.ApplicationList{}
			if err := r.List(ctx, appsList); err != nil {
				log.FromContext(ctx).Error(err, "Failed to list applications")
				return nil
			}

			var requests []reconcile2.Request
			for _, app := range appsList.Items {
				for _, appRepo := range app.Spec.Repositories {
					if appRepo.APIVersion == controllerRef.APIVersion && appRepo.Kind == controllerRef.Kind {
						if appRepo.Name == controllerRef.Name && appRepo.Namespace == ghRepoRef.Namespace {
							key := client.ObjectKey{Namespace: app.Namespace, Name: app.Name}
							requests = append(requests, reconcile2.Request{NamespacedName: key})
						}
					}
				}
			}
			return requests
		})).
		Complete(r)
}
