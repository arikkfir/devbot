package application

import (
	"context"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/secureworks/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"slices"
	"strings"
	"time"
)

var (
	ApplicationsFinalizer = "applications.finalizers." + apiv1.GroupVersion.Group
)

// Reconciler reconciles an Application object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	app := &apiv1.Application{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, errors.New("failed to get '%s'", req.NamespacedName, err)
		}
	}

	// Apply finalization if object is deleted
	if app.GetDeletionTimestamp() != nil {
		if slices.Contains(app.GetFinalizers(), ApplicationsFinalizer) {
			var finalizers []string
			for _, f := range app.GetFinalizers() {
				if f != ApplicationsFinalizer {
					finalizers = append(finalizers, f)
				}
			}
			app.SetFinalizers(finalizers)
			if err := r.Update(ctx, app); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it's missing
	if !slices.Contains(app.GetFinalizers(), ApplicationsFinalizer) {
		app.SetFinalizers(append(app.GetFinalizers(), ApplicationsFinalizer))
		if err := r.Update(ctx, app); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Initialize "Valid" condition
	if app.GetStatusConditionValid() == nil {
		app.SetStatusConditionValid(metav1.ConditionUnknown, apiv1.ReasonInitializing, "Initializing")
		if err := r.Status().Update(ctx, app); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Gather a distinct set of branch names from all repositories in this application
	distinctEnvNames := map[string]*apiv1.ApplicationEnvironment{}
	for _, repoRef := range app.Spec.Repositories {
		if repoRef.APIVersion == apiv1.GroupVersion.String() && repoRef.Kind == apiv1.GitHubRepositoryGVK.Kind {

			// This is a GitHub repository
			repo := &apiv1.GitHubRepository{}
			if err := r.Get(ctx, client.ObjectKey{Namespace: repoRef.Namespace, Name: repoRef.Name}, repo); err != nil {
				if client.IgnoreNotFound(err) == nil {

					// GitHubRepository does not exist - set status condition and requeue
					app.SetStatusConditionValid(metav1.ConditionFalse, apiv1.ReasonConfigError, fmt.Sprintf("GitHubRepository '%s/%s' not found", repoRef.Namespace, repoRef.Name))
					if err := r.Status().Update(ctx, app); err != nil {
						if apierrors.IsConflict(err) {
							return ctrl.Result{Requeue: true}, nil
						} else {
							return ctrl.Result{}, errors.New("failed to update status", err)
						}
					}
					return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil

				} else {

					// Failed querying for this GitHubRepository
					app.SetStatusConditionValid(metav1.ConditionUnknown, apiv1.ReasonInternalError, "Failed getting GitHubRepository '%s/%s': "+err.Error())
					if err := r.Status().Update(ctx, app); err != nil {
						if apierrors.IsConflict(err) {
							return ctrl.Result{Requeue: true}, nil
						} else {
							return ctrl.Result{}, errors.New("failed to update status", err)
						}
					}
					return ctrl.Result{}, errors.New("failed to get GitHubRepository '%s/%s'", repoRef.Namespace, repoRef.Name, err)

				}
			}

			// Get all GitHubRepositoryRef objects that belong to this repository
			refs := &apiv1.GitHubRepositoryRefList{}
			if err := k8s.GetOwnedChildrenByIndex(ctx, r.Client, repo, refs); err != nil {
				app.SetStatusConditionValid(metav1.ConditionUnknown, apiv1.ReasonInternalError, "Failed listing GitHubRepositoryRef objects: "+err.Error())
				if err := r.Status().Update(ctx, repo); err != nil {
					if apierrors.IsConflict(err) {
						return ctrl.Result{Requeue: true}, nil
					} else {
						return ctrl.Result{}, errors.New("failed to update status", err)
					}
				}
				return ctrl.Result{}, errors.New("failed listing GitHubRepositoryRef objects", err)
			}

			// Add all refs to the distinct set
			for _, ref := range refs.Items {
				branch := strings.TrimPrefix(ref.Spec.Ref, "refs/heads/")
				if _, exists := distinctEnvNames[branch]; !exists {
					distinctEnvNames[branch] = nil
				}
			}

		} else {

			// Unsupported type of repository
			app.SetStatusConditionValid(metav1.ConditionFalse, apiv1.ReasonConfigError, fmt.Sprintf("Unsupported repository object '%s.%s'", repoRef.Kind, repoRef.APIVersion))
			if err := r.Status().Update(ctx, app); err != nil {
				if apierrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				} else {
					return ctrl.Result{}, errors.New("failed to update status", err)
				}
			}
			return ctrl.Result{}, nil
		}
	}

	// List all ApplicationEnvironment objects owned by this application
	appEnvs := &apiv1.ApplicationEnvironmentList{}
	if err := k8s.GetOwnedChildrenByIndex(ctx, r.Client, app, appEnvs); err != nil {
		app.SetStatusConditionValid(metav1.ConditionUnknown, apiv1.ReasonInternalError, "Failed listing ApplicationEnvironment objects: "+err.Error())
		if err := r.Status().Update(ctx, app); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{}, errors.New("failed listing ApplicationEnvironment objects", err)
	}
	for _, appEnv := range appEnvs.Items {
		distinctEnvNames[appEnv.Spec.Branch] = &appEnv
	}

	// Ensure an ApplicationEnvironment object exists for each branch
	for envName, appEnv := range distinctEnvNames {
		if appEnv == nil {
			appEnv = &apiv1.ApplicationEnvironment{
				TypeMeta: metav1.TypeMeta{
					APIVersion: apiv1.GroupVersion.String(),
					Kind:       "ApplicationEnvironment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      stringsutil.RandomHash(7),
					Namespace: app.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         apiv1.GroupVersion.String(),
							Kind:               "Application",
							Name:               app.Name,
							UID:                app.UID,
							BlockOwnerDeletion: &[]bool{true}[0],
						},
					},
				},
				Spec: apiv1.ApplicationEnvironmentSpec{
					Branch: envName,
				},
			}
			if err := r.Create(ctx, appEnv); err != nil {
				app.SetStatusConditionValid(metav1.ConditionUnknown, apiv1.ReasonInternalError, "Failed creating ApplicationEnvironment object: "+err.Error())
				if err := r.Status().Update(ctx, app); err != nil {
					if apierrors.IsConflict(err) {
						return ctrl.Result{Requeue: true}, nil
					} else {
						return ctrl.Result{}, errors.New("failed to update status", err)
					}
				}
				return ctrl.Result{}, errors.New("failed creating ApplicationEnvironment object", err)
			}
		}
	}

	// If we got here, ensure the "Valid" condition is set to "True"
	if app.SetStatusConditionValidIfDifferent(metav1.ConditionTrue, apiv1.ReasonValid, "Valid") {
		if err := r.Status().Update(ctx, app); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// TODO: replace periodic requeue with a watch on the ApplicationEnvironment and GitHubRepositoryRef objects
	return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Application{}).
		Complete(r)
}
