package environment

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"slices"
)

var (
	Finalizer = "environments.finalizers." + apiv1.GroupVersion.Group
)

type (
	RepoDeploymentStatus struct {
		Deployment *apiv1.Deployment
		Branch     string
	}

	Reconciler struct {
		client.Client
		Scheme *runtime.Scheme
	}
)

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rec, result := k8s.NewReconciliation(ctx, r.Client, req, &apiv1.Environment{}, Finalizer, nil)
	if result != nil {
		return result.ToResultAndError()
	}

	// Finalize object if deleted
	if result := rec.FinalizeObjectIfDeleted(); result != nil {
		return result.ToResultAndError()
	}

	// Initialize object if not initialized
	if result := rec.InitializeObject(); result != nil {
		return result.ToResultAndError()
	}

	// Get controlling application
	app := &apiv1.Application{}
	if result := rec.GetRequiredController(app); result != nil {
		return result.ToResultAndError()
	}

	// Sync environment sources
	// Build a map of repository->branch
	// This map provides us with:
	// - the list of repositories that are participating in this environment
	// - the branch to deploy from each repository (either the preferred branch, the default branch, or no branch)
	repoBranches := make(map[apiv1.NamespacedRepositoryReference]RepoDeploymentStatus)
	for _, repoRef := range app.Spec.Repositories {
		nsRepoRef := repoRef.ToNamespacedRepositoryReference(app.Namespace)
		repoKey := repoRef.GetObjectKey(app.Namespace)
		if repoRef.IsGitHubRepository() {

			repo := &apiv1.GitHubRepository{}
			if err := r.Get(rec.Ctx, repoKey, repo); err != nil {
				if apierrors.IsNotFound(err) {
					rec.Object.Status.SetStaleDueToRepositoryNotFound("Repository '%s' not found", repoKey)
				} else if apierrors.IsForbidden(err) {
					rec.Object.Status.SetMaybeStaleDueToRepositoryNotAccessible("Repository '%s' is not accessible: %+v", repoKey, err)
				} else {
					rec.Object.Status.SetMaybeStaleDueToInternalError("Failed looking up repository '%s': %+v", repoKey, err)
				}
				if result := rec.UpdateStatus(); result != nil {
					return result.ToResultAndError()
				}
				return k8s.Requeue().ToResultAndError()
			} else if repo.Status.DefaultBranch == "" {
				rec.Object.Status.SetMaybeStaleDueToRepositoryNotReady("Repository '%s' is not ready: no default branch set", repoKey)
				if result := rec.UpdateStatus(); result != nil {
					return result.ToResultAndError()
				}
				return k8s.Requeue().ToResultAndError()
			}

			preferredBranchRefs := &apiv1.GitHubRepositoryRefList{}
			if err := r.List(rec.Ctx, preferredBranchRefs, k8s.OwnedBy(r.Scheme, repo), client.MatchingFields{"spec.ref": rec.Object.Spec.PreferredBranch}); err != nil {
				rec.Object.Status.SetMaybeStaleDueToInternalError("Failed looking up preferred ref for repository '%s': %+v", repoKey, err)
				if result := rec.UpdateStatus(); result != nil {
					return result.ToResultAndError()
				}
				return k8s.Requeue().ToResultAndError()
			} else if len(preferredBranchRefs.Items) == 0 {
				if repoRef.MissingBranchStrategy == apiv1.MissingBranchStrategyUseDefaultBranch {
					repoBranches[nsRepoRef] = RepoDeploymentStatus{Branch: repo.Status.DefaultBranch}
				} else if repoRef.MissingBranchStrategy == apiv1.MissingBranchStrategyIgnore {
					repoBranches[nsRepoRef] = RepoDeploymentStatus{Branch: ""}
				} else {
					rec.Object.Status.SetMaybeStaleDueToUnsupportedBranchStrategy("Repository '%s' has an unsupported missing branch strategy '%s'", repoKey, repoRef.MissingBranchStrategy)
					if result := rec.UpdateStatus(); result != nil {
						return result.ToResultAndError()
					}
					return k8s.Requeue().ToResultAndError()
				}
			} else {
				repoBranches[nsRepoRef] = RepoDeploymentStatus{Branch: rec.Object.Spec.PreferredBranch}
			}

		} else {
			rec.Object.Status.SetStaleDueToRepositoryNotSupported("Unsupported repository type '%s.%s' specified", repoRef.Kind, repoRef.APIVersion)
			if result := rec.UpdateStatus(); result != nil {
				return result.ToResultAndError()
			}
			return k8s.DoNotRequeue().ToResultAndError()
		}
	}

	// Fetch all our owned deployments
	depList := &apiv1.DeploymentList{}
	if err := r.List(rec.Ctx, depList, k8s.OwnedBy(r.Scheme, rec.Object)); err != nil {
		rec.Object.Status.SetMaybeStaleDueToInternalError("Failed to list deployments: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result.ToResultAndError()
		}
		return k8s.Requeue().ToResultAndError()
	}

	// Iterate the deployments and:
	// - remove deployments that are no longer participating in the application
	// - update deployments that are using the wrong branch
	for _, d := range depList.Items {

		if info, ok := repoBranches[d.Spec.Repository]; ok {

			if d.Spec.Branch != info.Branch {
				rec.Object.Status.SetMaybeStaleDueToDeploymentBranchOutOfSync("Deployment '%s' is out of sync: it should deploy branch '%s', but is set to deploy branch '%s'", d.Name, info.Branch, d.Spec.Branch)
				if result := rec.UpdateStatus(); result != nil {
					return result.ToResultAndError()
				}

				d.Spec.Branch = info.Branch
				if err := r.Update(rec.Ctx, &d); err != nil {
					if apierrors.IsNotFound(err) {
						return k8s.Requeue().ToResultAndError()
					} else if apierrors.IsConflict(err) {
						return k8s.Requeue().ToResultAndError()
					} else {
						return k8s.RequeueDueToError(err).ToResultAndError()
					}
				}
			}
			repoBranches[d.Spec.Repository] = RepoDeploymentStatus{Deployment: &d, Branch: info.Branch}

		} else {

			if err := r.Delete(rec.Ctx, &d); err != nil {
				if apierrors.IsNotFound(err) {
					// Ignore
				} else if apierrors.IsConflict(err) {
					return k8s.Requeue().ToResultAndError()
				} else {
					return k8s.RequeueDueToError(err).ToResultAndError()
				}
			}

		}
	}

	// Iterate participating repositories and make sure each one has a corresponding deployment
	for repoRef, deploymentStatus := range repoBranches {
		if deploymentStatus.Deployment == nil {
			deployment := &apiv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:            strings.RandomHash(7),
					Namespace:       rec.Object.Namespace,
					OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(rec.Object, apiv1.EnvironmentGVK)},
				},
				Spec: apiv1.DeploymentSpec{
					Repository: repoRef,
					Branch:     deploymentStatus.Branch,
				},
			}
			if err := r.Create(rec.Ctx, deployment); err != nil {
				rec.Object.Status.SetMaybeStaleDueToInternalError("Failed to create deployment '%s': %+v", deployment.Name, err)
				if result := rec.UpdateStatus(); result != nil {
					return result.ToResultAndError()
				}
			}
			return k8s.Requeue().ToResultAndError()
		}
	}

	// Remove stale condition if we got here
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.RepositoryNotFound, apiv1.RepositoryNotAccessible, apiv1.InternalError, apiv1.RepositoryNotReady, apiv1.UnsupportedBranchStrategy, apiv1.RepositoryNotSupported, apiv1.DeploymentBranchOutOfSync)
	if result := rec.UpdateStatus(); result != nil {
		return result.ToResultAndError()
	}

	// Mark as stale if any deployment is stale; current otherwise
	rec.Object.Status.SetCurrent()
	for _, deployment := range depList.Items {
		if deployment.Status.IsStale() {
			rec.Object.Status.SetStaleDueToDeploymentsAreStale("One or more deployments are stale")
			break
		}
	}
	if result := rec.UpdateStatus(); result != nil {
		return result.ToResultAndError()
	}

	// Done
	return k8s.DoNotRequeue().ToResultAndError()
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Environment{}).
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
		Watches(&apiv1.GitHubRepository{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
			ghRepo := obj.(*apiv1.GitHubRepository)
			ghRepoKey := client.ObjectKeyFromObject(ghRepo)

			appsList := &apiv1.ApplicationList{}
			if err := r.List(ctx, appsList); err != nil {
				log.FromContext(ctx).Error(err, "Failed to list applications")
				return nil
			}

			var requests []ctrl.Request
			for _, app := range appsList.Items {
				hasRepo := func(r apiv1.ApplicationSpecRepository) bool {
					return r.APIVersion == ghRepo.APIVersion && r.Kind == ghRepo.Kind && ghRepoKey == r.GetObjectKey(app.Namespace)
				}

				if slices.ContainsFunc(app.Spec.Repositories, hasRepo) {
					envsList := &apiv1.EnvironmentList{}
					if err := r.List(ctx, envsList, client.InNamespace(app.Namespace), k8s.OwnedBy(r.Scheme, &app)); err != nil {
						log.FromContext(ctx).Error(err, "Failed to list environments")
						return nil
					}
					for _, item := range envsList.Items {
						requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&item)})
					}
				}
			}
			return requests
		})).
		Watches(&apiv1.GitHubRepositoryRef{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
			ghRepoRef := obj.(*apiv1.GitHubRepositoryRef)
			ctrlRef := metav1.GetControllerOf(ghRepoRef)
			if ctrlRef == nil {
				return nil
			}
			ghRepoKey := client.ObjectKey{Namespace: ghRepoRef.Namespace, Name: ctrlRef.Name}

			appsList := &apiv1.ApplicationList{}
			if err := r.List(ctx, appsList); err != nil {
				log.FromContext(ctx).Error(err, "Failed to list applications")
				return nil
			}

			var requests []reconcile.Request
			for _, app := range appsList.Items {
				hasRepo := func(r apiv1.ApplicationSpecRepository) bool {
					return r.APIVersion == ctrlRef.APIVersion && r.Kind == ctrlRef.Kind && ghRepoKey == r.GetObjectKey(app.Namespace)
				}
				if slices.ContainsFunc(app.Spec.Repositories, hasRepo) {
					envsList := &apiv1.EnvironmentList{}
					if err := r.List(ctx, envsList, client.InNamespace(app.Namespace), k8s.OwnedBy(r.Scheme, &app)); err != nil {
						log.FromContext(ctx).Error(err, "Failed to list environments")
						return nil
					}
					for _, item := range envsList.Items {
						requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&item)})
					}
				}
			}
			return requests
		})).
		Complete(r)
}
