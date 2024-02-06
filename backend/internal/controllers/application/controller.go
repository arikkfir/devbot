package application

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/secureworks/errors"
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
	rec, result := k8s.NewReconciliation(ctx, r.Client, req, &apiv1.Application{}, Finalizer, nil)
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

	// Create missing environments objects based on current branches in participating repositories
	// Fetch existing ref objects
	envsList := &apiv1.EnvironmentList{}
	if err := r.List(rec.Ctx, envsList, k8s.OwnedBy(r.Scheme, rec.Object)); err != nil {
		return k8s.RequeueDueToError(errors.New("failed listing owned objects: %w", err)).ToResultAndError()
	}

	// Create a map of all environments by their preferred branch
	envByBranch := make(map[string]*apiv1.Environment)
	for i, item := range envsList.Items {
		envByBranch[item.Spec.PreferredBranch] = &envsList.Items[i]
	}
	var namesOfEnvsToRetain []string

	// Ensure all branches from all participating repositories are mapped to an environment
	for _, repoRef := range rec.Object.Spec.Repositories {
		repoClientKey := repoRef.GetObjectKey(rec.Object.Namespace)
		if repoRef.IsGitHubRepository() {

			// This is a GitHub repository
			repo := &apiv1.GitHubRepository{}
			if err := r.Get(rec.Ctx, repoClientKey, repo); err != nil {
				if apierrors.IsNotFound(err) {
					rec.Object.Status.SetStaleDueToRepositoryNotFound("GitHub repository '%s' not found", repoClientKey)
					if result := rec.UpdateStatus(); result != nil {
						return result.ToResultAndError()
					}
					return k8s.Requeue().ToResultAndError()
				} else if apierrors.IsForbidden(err) {
					rec.Object.Status.SetMaybeStaleDueToRepositoryNotAccessible("GitHub repository '%s' is not accessible: %+v", repoClientKey, err)
					if result := rec.UpdateStatus(); result != nil {
						return result.ToResultAndError()
					}
					return k8s.Requeue().ToResultAndError()
				} else {
					rec.Object.Status.SetMaybeStaleDueToInternalError("Failed looking up GitHub repository '%s': %+v", repoClientKey, err)
					if result := rec.UpdateStatus(); result != nil {
						return result.ToResultAndError()
					}
					return k8s.Requeue().ToResultAndError()
				}
			}

			// Get all GitHubRepositoryRef objects that belong to this repository
			refs := &apiv1.GitHubRepositoryRefList{}
			if err := r.List(rec.Ctx, refs, k8s.OwnedBy(r.Scheme, repo)); err != nil {
				rec.Object.Status.SetMaybeStaleDueToInternalError("Failed looking up refs for GitHub repository '%s': %+v", repoClientKey, err)
				if result := rec.UpdateStatus(); result != nil {
					return result.ToResultAndError()
				}
				return k8s.Requeue().ToResultAndError()
			}

			// For every repository branch found, ensure there's an environment for it
			for _, ref := range refs.Items {
				namesOfEnvsToRetain = append(namesOfEnvsToRetain, ref.Spec.Ref)
				if _, ok := envByBranch[ref.Spec.Ref]; !ok {
					env := &apiv1.Environment{
						ObjectMeta: metav1.ObjectMeta{
							Name:            strings.RandomHash(7),
							Namespace:       rec.Object.Namespace,
							OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(rec.Object, apiv1.ApplicationGVK)},
						},
						Spec: apiv1.EnvironmentSpec{PreferredBranch: ref.Spec.Ref},
					}
					if err := r.Create(rec.Ctx, env); err != nil {
						rec.Object.Status.SetMaybeStaleDueToInternalError("Failed creating environment for branch '%s': %+v", ref.Spec.Ref, err)
						if result := rec.UpdateStatus(); result != nil {
							return result.ToResultAndError()
						}
						return k8s.Requeue().ToResultAndError()
					}
				}
			}

		} else {
			rec.Object.Status.SetInvalidDueToRepositoryNotSupported("Unsupported repository type '%s.%s' specified", repoRef.Kind, repoRef.APIVersion)
			rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
			if result := rec.UpdateStatus(); result != nil {
				return result.ToResultAndError()
			}
			return k8s.DoNotRequeue().ToResultAndError()
		}
	}

	// Delete stale environments
	for _, env := range envsList.Items {
		if !slices.Contains(namesOfEnvsToRetain, env.Spec.PreferredBranch) {
			if err := r.Delete(rec.Ctx, &env); err != nil {
				if apierrors.IsNotFound(err) {
					continue
				} else {
					rec.Object.Status.SetStaleDueToInternalError("Failed deleting environment '%s': %+v", env.Name, err)
					if result := rec.UpdateStatus(); result != nil {
						return result.ToResultAndError()
					}
					return k8s.Requeue().ToResultAndError()
				}
			}
		}
	}

	// Reset conditions to valid state since all has passed successfully
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.RepositoryNotFound, apiv1.RepositoryNotAccessible, apiv1.InternalError, apiv1.Invalid)
	rec.Object.Status.SetValidIfInvalidDueToAnyOf(apiv1.RepositoryNotSupported)
	if result := rec.UpdateStatus(); result != nil {
		return result.ToResultAndError()
	}

	// Done (we're watching repositories & refs, so no need to requeue)
	return k8s.DoNotRequeue().ToResultAndError()
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Application{}).
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
				hasWatchedRepository := func(r apiv1.ApplicationSpecRepository) bool {
					return r.APIVersion == ghRepo.APIVersion && r.Kind == ghRepo.Kind && ghRepoKey == r.GetObjectKey(app.Namespace)
				}
				if slices.ContainsFunc(app.Spec.Repositories, hasWatchedRepository) {
					requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&app)})
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
				hasRepositoryOfWatchedRef := func(r apiv1.ApplicationSpecRepository) bool {
					return r.APIVersion == ctrlRef.APIVersion && r.Kind == ctrlRef.Kind && ghRepoKey == r.GetObjectKey(app.Namespace)
				}
				if slices.ContainsFunc(app.Spec.Repositories, hasRepositoryOfWatchedRef) {
					requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&app)})
				}
			}
			return requests
		})).
		Complete(r)
}