package application

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/config"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/secureworks/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"slices"
)

var (
	Finalizer = "applications.finalizers." + apiv1.GroupVersion.Group
)

// Reconciler reconciles an Application object
type Reconciler struct {
	client.Client
	Config config.CommandConfig
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.executeReconciliation(ctx, req).ToResultAndError()
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) executeReconciliation(ctx context.Context, req ctrl.Request) *k8s.Result {
	rec, result := k8s.NewReconciliation(ctx, r.Config, r.Client, req, &apiv1.Application{}, Finalizer, nil)
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

	// Create missing environments objects based on current branches in participating repositories
	// Fetch existing ref objects
	envsList := &apiv1.EnvironmentList{}
	if err := r.List(rec.Ctx, envsList, k8s.OwnedBy(r.Scheme, rec.Object)); err != nil {
		return k8s.RequeueDueToError(errors.New("failed listing owned objects: %w", err))
	}

	// Create a map of all environments by their preferred branch
	existingEnvironmentsByBranch := make(map[string]*apiv1.Environment)
	for i, item := range envsList.Items {
		existingEnvironmentsByBranch[item.Spec.PreferredBranch] = &envsList.Items[i]
	}
	var namesOfEnvsToRetain []string

	// Setup branch regular expressions
	var allowedBranches []*regexp.Regexp
	for _, branchExpr := range rec.Object.Spec.Branches {
		re, err := regexp.Compile(branchExpr)
		if err != nil {
			rec.Object.Status.SetInvalidDueToInvalidBranchSpecification("Invalid branch specification: %s", branchExpr)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
		} else {
			allowedBranches = append(allowedBranches, re)
		}
	}
	rec.Object.Status.SetValidIfInvalidDueToAnyOf(apiv1.InvalidBranchSpecification)
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Ensure all branches from all participating repositories are mapped to an environment
	for _, repoRef := range rec.Object.Spec.Repositories {
		repoKey := repoRef.GetObjectKey(rec.Object.Namespace)

		// Fetch repository
		repo := &apiv1.Repository{}
		if err := r.Get(rec.Ctx, repoKey, repo); err != nil {
			if apierrors.IsNotFound(err) {
				rec.Object.Status.SetStaleDueToRepositoryNotFound("Repository '%s' not found", repoKey)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
				return k8s.Requeue()
			} else if apierrors.IsForbidden(err) {
				rec.Object.Status.SetMaybeStaleDueToRepositoryNotAccessible("Repository '%s' is not accessible: %+v", repoKey, err)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
				return k8s.Requeue()
			} else {
				rec.Object.Status.SetMaybeStaleDueToInternalError("Failed looking up repository '%s': %+v", repoKey, err)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
				return k8s.Requeue()
			}
		}

		// For every repository branch, ensure there's an environment for it
		for branch := range repo.Status.Revisions {

			// Skip branches that don't match any of the allowed branch expressions
			if len(allowedBranches) > 0 {
				matches := false
				for _, branchRE := range allowedBranches {
					if branchRE.MatchString(branch) {
						matches = true
						break
					}
				}
				if !matches {
					continue
				}
			}

			// Create an environment for this branch, if one does not yet exist
			if _, ok := existingEnvironmentsByBranch[branch]; !ok {
				env := &apiv1.Environment{
					ObjectMeta: metav1.ObjectMeta{
						Name:            strings.RandomHash(7),
						Namespace:       rec.Object.Namespace,
						OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(rec.Object, apiv1.ApplicationGVK)},
					},
					Spec: apiv1.EnvironmentSpec{PreferredBranch: branch},
				}
				if err := r.Create(rec.Ctx, env); err != nil {
					rec.Object.Status.SetMaybeStaleDueToInternalError("Failed creating environment for branch '%s': %+v", branch, err)
					if result := rec.UpdateStatus(); result != nil {
						return result
					}
					return k8s.Requeue()
				}
				existingEnvironmentsByBranch[branch] = env
			}

			// Remember to keep this environment, since there's an active branch for it
			namesOfEnvsToRetain = append(namesOfEnvsToRetain, branch)
		}
	}

	// Prune environments with no matching branch names in any of the app's repositories
	for _, env := range envsList.Items {
		if !slices.Contains(namesOfEnvsToRetain, env.Spec.PreferredBranch) {
			if err := r.Delete(rec.Ctx, &env); err != nil {
				if !apierrors.IsNotFound(err) {
					rec.Object.Status.SetStaleDueToInternalError("Failed deleting environment '%s': %+v", env.Name, err)
					if result := rec.UpdateStatus(); result != nil {
						return result
					}
					return k8s.Requeue()
				}
			}
		}
	}

	// Mark as stale if any deployment is stale; current otherwise
	rec.Object.Status.SetCurrent()
	for _, environment := range envsList.Items {
		if environment.Status.IsStale() {
			rec.Object.Status.SetStaleDueToEnvironmentsAreStale("One or more environments are stale")
		}
	}
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Done (we're watching repositories, so no need to requeue)
	return k8s.DoNotRequeue()
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Application{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				// Only reconcile if the generation has changed
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			},
		})).
		Watches(&apiv1.Environment{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			environment := obj.(*apiv1.Environment)
			appRef := metav1.GetControllerOf(environment)
			if appRef == nil {
				return nil
			}
			return []reconcile.Request{{NamespacedName: client.ObjectKey{Namespace: environment.Namespace, Name: appRef.Name}}}
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
				for _, appRepo := range app.Spec.Repositories {
					if appRepo.GetObjectKey(app.Namespace) == repoKey {
						// This app uses the repository that triggered this watcher
						requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&app)})
					}
				}
			}
			return requests
		})).
		Complete(r)
}
