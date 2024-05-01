package environment

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/config"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var (
	Finalizer = "environments.finalizers." + apiv1.GroupVersion.Group
)

type (
	RepoDeploymentStatus struct {
		Deployment *apiv1.Deployment
		BranchRef  *apiv1.NamespacedReference
	}

	Reconciler struct {
		client.Client
		Config config.CommandConfig
		Scheme *runtime.Scheme
	}
)

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.executeReconciliation(ctx, req).ToResultAndError()
}

func (r *Reconciler) executeReconciliation(ctx context.Context, req ctrl.Request) *k8s.Result {
	rec, result := k8s.NewReconciliation(ctx, r.Config, r.Client, req, &apiv1.Environment{}, Finalizer, nil)
	if result != nil {
		return result
	}

	// Finalize object if deleted
	if result := rec.FinalizeObjectIfDeleted(); result != nil {
		return result
	}

	// Initialize object if not initialized
	if result := rec.InitializeObject(); result != nil {
		return result
	}

	// Get controlling application
	app := &apiv1.Application{}
	if result := rec.GetRequiredController(app); result != nil {
		return result
	}

	// Get all controlled Deployment objects
	deployments := &apiv1.DeploymentList{}
	if err := r.List(rec.Ctx, deployments, k8s.OwnedBy(r.Client.Scheme(), rec.Object)); err != nil {
		rec.Object.Status.SetMaybeStaleDueToInternalError("Failed to list deployments: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// For each participating repository, verify that there's a corresponding Deployment object
	for _, repoRef := range app.Spec.Repositories {
		repoKey := repoRef.GetObjectKey(app.Namespace)

		found := false
		for _, d := range deployments.Items {
			if d.Spec.Repository.GetObjectKey(d.Namespace) == repoKey {
				found = true
				break
			}
		}

		if !found {
			d := &apiv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      strings.RandomHash(7),
					Namespace: rec.Object.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(rec.Object, apiv1.EnvironmentGVK),
					},
				},
				Spec: apiv1.DeploymentSpec{Repository: apiv1.DeploymentRepositoryReference{
					Name:      repoKey.Name,
					Namespace: repoKey.Namespace,
				}},
			}
			if err := r.Create(rec.Ctx, d); err != nil {
				rec.Object.Status.SetStaleDueToFailedCreatingDeployment("Failed to create deployment for repository '%s': %+v", repoKey, err)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
				return k8s.Requeue()
			}
		}
	}

	// Prune deployments for repositories that are no longer participating in the application
	for _, d := range deployments.Items {

		found := false
		for _, appRepoRef := range app.Spec.Repositories {
			appRepoKey := appRepoRef.GetObjectKey(app.Namespace)
			if d.Spec.Repository.GetObjectKey(d.Namespace) == appRepoKey {
				found = true
				break
			}
		}

		if !found {
			if err := r.Delete(rec.Ctx, &d); err != nil {
				if apierrors.IsNotFound(err) {
					// Ignore
				} else if apierrors.IsConflict(err) {
					return k8s.Requeue()
				} else {
					rec.Object.Status.SetStaleDueToFailedDeletingDeployment("Failed to delete deployment '%s/%s': %+v", d.Namespace, d.Name, err)
					if result := rec.UpdateStatus(); result != nil {
						return result
					}
					return k8s.Requeue()
				}
			}
		}
	}

	// Remove stale condition if we got here
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.FailedCreatingDeployment, apiv1.FailedDeletingDeployment, apiv1.InternalError)
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Mark as stale if any deployment is stale; current otherwise
	for _, deployment := range deployments.Items {
		if deployment.Status.IsStale() {
			rec.Object.Status.SetStaleDueToDeploymentsAreStale("One or more deployments are stale")
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.RequeueAfter(30 * time.Second)
		}
	}

	// Mark as current if we got this far
	rec.Object.Status.SetCurrent()
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Done
	return k8s.DoNotRequeue()
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Environment{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				// Only reconcile if the generation has changed
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			},
		})).
		Watches(&apiv1.Application{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			app := obj.(*apiv1.Application)

			envsList := &apiv1.EnvironmentList{}
			if err := r.List(ctx, envsList, k8s.OwnedBy(r.Scheme, app)); err != nil {
				log.FromContext(ctx).Error(err, "Failed to list environments")
				return nil
			}

			var requests []reconcile.Request
			for _, env := range envsList.Items {
				requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&env)})
			}
			return requests
		})).
		Watches(&apiv1.Deployment{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			deployment := obj.(*apiv1.Deployment)
			envRef := metav1.GetControllerOf(deployment)
			if envRef == nil {
				return nil
			}
			return []reconcile.Request{{NamespacedName: client.ObjectKey{Namespace: deployment.Namespace, Name: envRef.Name}}}
		})).
		Watches(&apiv1.Repository{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
			repo := obj.(*apiv1.Repository)
			repoKey := client.ObjectKeyFromObject(repo)

			appsList := &apiv1.ApplicationList{}
			if err := r.List(ctx, appsList); err != nil {
				log.FromContext(ctx).Error(err, "Failed to list applications")
				return nil
			}

			var requests []ctrl.Request
			for _, app := range appsList.Items {
				repositoryParticipatesInThisApp := false
				for _, appRepo := range app.Spec.Repositories {
					if appRepo.GetObjectKey(app.Namespace) == repoKey {
						repositoryParticipatesInThisApp = true
						break
					}
				}

				if repositoryParticipatesInThisApp {
					environmentsList := &apiv1.EnvironmentList{}
					if err := r.List(ctx, environmentsList, k8s.OwnedBy(r.Client.Scheme(), &app)); err != nil {
						log.FromContext(ctx).Error(err, "Failed to list applications")
						return nil
					}
					for _, env := range environmentsList.Items {
						requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&env)})
					}
				}
			}
			return requests
		})).
		Complete(r)
}
