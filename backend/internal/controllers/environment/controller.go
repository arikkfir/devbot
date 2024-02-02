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
	"time"
)

var (
	Finalizer = "environments.finalizers." + apiv1.GroupVersion.Group
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *Reconciler) syncSources(rec *k8s.Reconciliation[*apiv1.Environment], app *apiv1.Application) *k8s.Result {
	type RepoDeploymentStatus struct {
		Deployment *apiv1.Deployment
		Branch     string
	}

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
					if rec.Object.Status.SetStaleDueToRepositoryNotFound("Repository '%s' not found", repoKey) {
						return rec.UpdateStatus(k8s.WithStrategy(k8s.Requeue))
					} else {
						return k8s.Requeue()
					}
				} else if apierrors.IsForbidden(err) {
					if rec.Object.Status.SetMaybeStaleDueToRepositoryNotAccessible("Repository '%s' is not accessible: %+v", repoKey, err) {
						return rec.UpdateStatus(k8s.WithStrategy(k8s.Requeue))
					} else {
						return k8s.Requeue()
					}
				} else {
					if rec.Object.Status.SetMaybeStaleDueToInternalError("Failed looking up repository '%s': %+v", repoKey, err) {
						return rec.UpdateStatus(k8s.WithStrategy(k8s.Requeue))
					} else {
						return k8s.Requeue()
					}
				}
			} else if repo.Status.DefaultBranch == "" {
				if rec.Object.Status.SetMaybeStaleDueToRepositoryNotReady("Repository '%s' is not ready: no default branch set", repoKey) {
					return rec.UpdateStatus(k8s.WithStrategy(k8s.Requeue))
				} else {
					return k8s.Requeue()
				}
			}

			preferredBranchRefs := &apiv1.GitHubRepositoryRefList{}
			if err := r.List(rec.Ctx, preferredBranchRefs, k8s.OwnedBy(r.Scheme, repo), client.MatchingFields{"spec.ref": rec.Object.Spec.PreferredBranch}); err != nil {
				if rec.Object.Status.SetMaybeStaleDueToInternalError("Failed looking up preferred ref for repository '%s': %+v", repoKey, err) {
					return rec.UpdateStatus(k8s.WithStrategy(k8s.Requeue))
				} else {
					return k8s.Requeue()
				}
			} else if len(preferredBranchRefs.Items) == 0 {
				if repoRef.MissingBranchStrategy == apiv1.MissingBranchStrategyUseDefaultBranch {
					repoBranches[nsRepoRef] = RepoDeploymentStatus{Branch: repo.Status.DefaultBranch}
				} else if repoRef.MissingBranchStrategy == apiv1.MissingBranchStrategyIgnore {
					repoBranches[nsRepoRef] = RepoDeploymentStatus{Branch: ""}
				} else {
					if rec.Object.Status.SetMaybeStaleDueToUnsupportedBranchStrategy("Repository '%s' has an unsupported missing branch strategy '%s'", repoKey, repoRef.MissingBranchStrategy) {
						return rec.UpdateStatus(k8s.WithStrategy(k8s.DoNotRequeue))
					} else {
						return k8s.DoNotRequeue()
					}
				}
			} else {
				repoBranches[nsRepoRef] = RepoDeploymentStatus{Branch: rec.Object.Spec.PreferredBranch}
			}

		} else {
			if rec.Object.Status.SetStaleDueToRepositoryNotSupported("Unsupported repository type '%s.%s' specified", repoRef.Kind, repoRef.APIVersion) {
				return rec.UpdateStatus(k8s.WithStrategy(k8s.DoNotRequeue))
			} else {
				return k8s.DoNotRequeue()
			}
		}
	}

	// Iterate all our deployments and:
	// - remove deployments that are no longer participating in the application
	// - update deployments that are using the wrong branch
	depList := &apiv1.DeploymentList{}
	if err := r.List(rec.Ctx, depList, k8s.OwnedBy(r.Scheme, rec.Object)); err != nil {
		if rec.Object.Status.SetMaybeStaleDueToInternalError("Failed to list deployments: %+v", err) {
			return rec.UpdateStatus(k8s.WithStrategy(k8s.Requeue))
		} else {
			return k8s.Requeue()
		}
	}
	for _, d := range depList.Items {
		if info, ok := repoBranches[d.Spec.Repository]; ok {
			if d.Spec.Branch != info.Branch {
				d.Spec.Branch = info.Branch
				if err := r.Update(rec.Ctx, &d); err != nil {
					if apierrors.IsNotFound(err) {
						return k8s.Requeue()
					} else if apierrors.IsConflict(err) {
						return k8s.Requeue()
					} else {
						if rec.Object.Status.SetMaybeStaleDueToInternalError("Failed to update deployment '%s': %+v", d.Name, err) {
							return rec.UpdateStatus(k8s.WithStrategy(k8s.Requeue))
						} else {
							return k8s.Requeue()
						}
					}
				}
			}
			repoBranches[d.Spec.Repository] = RepoDeploymentStatus{Deployment: &d, Branch: info.Branch}
		} else {
			if err := r.Delete(rec.Ctx, &d); err != nil {
				if apierrors.IsNotFound(err) {
					// Ignore
				} else if apierrors.IsConflict(err) {
					return k8s.Requeue()
				} else {
					if rec.Object.Status.SetMaybeStaleDueToInternalError("Failed to delete deployment '%s': %+v", d.Name, err) {
						return rec.UpdateStatus(k8s.WithStrategy(k8s.Requeue))
					} else {
						return k8s.Requeue()
					}
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
				if rec.Object.Status.SetMaybeStaleDueToInternalError("Failed to create deployment '%s': %+v", deployment.Name, err) {
					return rec.UpdateStatus(k8s.WithStrategy(k8s.Requeue))
				} else {
					return k8s.Requeue()
				}
			}
		}
	}

	if rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.RepositoryNotFound, apiv1.RepositoryNotAccessible, apiv1.InternalError, apiv1.RepositoryNotReady, apiv1.UnsupportedBranchStrategy, apiv1.RepositoryNotSupported) {
		return rec.UpdateStatus(k8s.WithStrategy(k8s.Continue))
	} else {
		return k8s.Continue()
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rec, result := k8s.NewReconciliation(ctx, r.Client, req, &apiv1.Environment{}, Finalizer, nil)
	if result != nil {
		return result.Return()
	}

	// Finalize object if deleted
	if result := rec.FinalizeObjectIfDeleted(); result != nil {
		return result.Return()
	}

	// Initialize object if not initialized
	if result := rec.InitializeObject(); result != nil {
		return result.Return()
	}

	// Get controlling application
	app := &apiv1.Application{}
	if result := rec.GetRequiredController(app); result != nil {
		return result.Return()
	}

	// Sync environment sources
	if result := r.syncSources(rec, app); result != nil {
		return result.Return()
	}

	// Done
	// TODO: replace RequeueAfter with watching over our repositories & branches
	return k8s.RequeueAfter(1 * time.Minute).Return()
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
		Complete(r)
}
