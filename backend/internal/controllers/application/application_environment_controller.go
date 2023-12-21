package application

import (
	"context"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/secureworks/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"slices"
	"strings"
)

var (
	ApplicationEnvironmentsFinalizer = "environments.applications.finalizers." + apiv1.GroupVersion.Group
)

type EnvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *EnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	env := &apiv1.ApplicationEnvironment{}
	if err := r.Get(ctx, req.NamespacedName, env); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, errors.New("failed to get '%s'", req.NamespacedName, err)
		}
	}

	// Apply finalization if object is deleted
	if env.GetDeletionTimestamp() != nil {
		if slices.Contains(env.GetFinalizers(), ApplicationEnvironmentsFinalizer) {
			var finalizers []string
			for _, f := range env.GetFinalizers() {
				if f != ApplicationEnvironmentsFinalizer {
					finalizers = append(finalizers, f)
				}
			}
			env.SetFinalizers(finalizers)
			if err := r.Update(ctx, env); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it's missing
	if !slices.Contains(env.GetFinalizers(), ApplicationEnvironmentsFinalizer) {
		env.SetFinalizers(append(env.GetFinalizers(), ApplicationEnvironmentsFinalizer))
		if err := r.Update(ctx, env); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Initialize "Current" condition
	if env.GetStatusConditionValid() == nil {
		env.SetStatusConditionValid(metav1.ConditionUnknown, apiv1.ReasonInitializing, "Initializing")
		if err := r.Status().Update(ctx, env); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Find our owner application
	app := &apiv1.Application{}
	if err := k8s.GetFirstOwnerOfType(ctx, r.Client, env, apiv1.GroupVersion.String(), "Application", app); err != nil {
		env.SetStatusConditionValid(metav1.ConditionUnknown, apiv1.ReasonConfigError, "Owner not found: "+err.Error())
		if err := r.Status().Update(ctx, env); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{}, errors.New("failed to get owner", err)
	}

	// Find all our owned deployments
	depList := &apiv1.DeploymentList{}
	if err := k8s.GetOwnedChildrenByIndex(ctx, r.Client, env, depList); err != nil {
		env.SetStatusConditionValid(metav1.ConditionUnknown, apiv1.ReasonInternalError, "Failed listing deployments: "+err.Error())
		if err := r.Status().Update(ctx, env); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{}, errors.New("failed listing deployments", err)
	}

	// Start off with all Deployment objects destined for deletion
	// Every deployment in the process that we find is still needed will be removed from this list
	var deploymentsToDelete []*apiv1.Deployment
	for _, deployment := range depList.Items {
		d := deployment
		deploymentsToDelete = append(deploymentsToDelete, &d)
	}

	// Iterate repositories, ensuring we have a corresponding appropriate Deployment for each one
	for _, repoRef := range app.Spec.Repositories {

		if repoRef.APIVersion == apiv1.GroupVersion.String() && repoRef.Kind == apiv1.GitHubRepositoryGVK.Kind {

			// Find the GitHubRepository object
			repo := &apiv1.GitHubRepository{}
			if err := r.Get(ctx, client.ObjectKey{Namespace: repoRef.Namespace, Name: repoRef.Name}, repo); err != nil {
				if client.IgnoreNotFound(err) == nil {
					env.SetStatusConditionValid(metav1.ConditionFalse, apiv1.ReasonConfigError, fmt.Sprintf("GitHub repository object '%s/%s' not found", repoRef.Namespace, repoRef.Name))
					if err := r.Status().Update(ctx, env); err != nil {
						if apierrors.IsConflict(err) {
							return ctrl.Result{Requeue: true}, nil
						} else {
							return ctrl.Result{}, errors.New("failed to update status", err)
						}
					}
					return ctrl.Result{}, errors.New("missing GitHub repository object: %s/%s", repoRef.Namespace, repoRef.Name)
				} else {
					env.SetStatusConditionValid(metav1.ConditionUnknown, apiv1.ReasonInternalError, fmt.Sprintf("Failed getting GitHub repository object '%s/%s': %s", repoRef.Namespace, repoRef.Name, err.Error()))
					if err := r.Status().Update(ctx, env); err != nil {
						if apierrors.IsConflict(err) {
							return ctrl.Result{Requeue: true}, nil
						} else {
							return ctrl.Result{}, errors.New("failed to update status", err)
						}
					}
					return ctrl.Result{}, errors.New("failed getting GitHub repository object: %s/%s", repoRef.Namespace, repoRef.Name, err)
				}
			}

			// Find the GitHubRepositoryRef owned by current GitHubRepository referencing the current branch
			refs := &apiv1.GitHubRepositoryRefList{}
			var ref *apiv1.GitHubRepositoryRef
			if err := k8s.GetOwnedChildrenByIndex(ctx, r.Client, repo, refs); err != nil {
				env.SetStatusConditionValid(metav1.ConditionUnknown, apiv1.ReasonInternalError, fmt.Sprintf("Failed listing GitHub repository refs: %s", err.Error()))
				if err := r.Status().Update(ctx, env); err != nil {
					if apierrors.IsConflict(err) {
						return ctrl.Result{Requeue: true}, nil
					} else {
						return ctrl.Result{}, errors.New("failed to update status", err)
					}
				}
				return ctrl.Result{}, errors.New("failed listing GitHub repository refs", err)
			}
			for _, item := range refs.Items {
				if item.Spec.Ref == "refs/heads/"+env.Spec.Branch {
					ref = &item
					break
				}
			}

			// If GitHubRepositoryRef not found, consult with the owning application's strategy
			if ref == nil {

				if strings.ToLower(app.Spec.Strategy.Missing) == "ignore" {

					// If strategy is "ignore", we don't need to do anything (deployment will be deleted if it exists)
					continue

				} else if strings.ToLower(app.Spec.Strategy.Missing) == "default" {

					// If strategy is "default", we need to find the default branch from the owning application

					// If no default branch is specified, fail and signal a configuration error
					if app.Spec.Strategy.DefaultBranch == "" {
						env.SetStatusConditionValid(metav1.ConditionFalse, apiv1.ReasonConfigError, fmt.Sprintf("Default branch needed but missing in application"))
						if err := r.Status().Update(ctx, env); err != nil {
							if apierrors.IsConflict(err) {
								return ctrl.Result{Requeue: true}, nil
							} else {
								return ctrl.Result{}, errors.New("failed to update status", err)
							}
						}
						return ctrl.Result{}, errors.New("default branch needed but missing in application")
					}

					// Find the GitHubRepositoryRef owned by current GitHubRepository referencing the default branch
					for _, item := range refs.Items {
						if item.Spec.Ref == "refs/heads/"+app.Spec.Strategy.DefaultBranch {
							ref = &item
							break
						}
					}

					// If still not found, fail and signal a configuration error
					if ref == nil {
						env.SetStatusConditionValid(metav1.ConditionFalse, apiv1.ReasonConfigError, fmt.Sprintf("Default branch needed but does not exist"))
						if err := r.Status().Update(ctx, env); err != nil {
							if apierrors.IsConflict(err) {
								return ctrl.Result{Requeue: true}, nil
							} else {
								return ctrl.Result{}, errors.New("failed to update status", err)
							}
						}
						return ctrl.Result{}, errors.New("default branch needed but does not exist")
					}

				} else {
					env.SetStatusConditionValid(metav1.ConditionFalse, apiv1.ReasonConfigError, fmt.Sprintf("Unsupported application strategy: %s", app.Spec.Strategy.Missing))
					if err := r.Status().Update(ctx, env); err != nil {
						if apierrors.IsConflict(err) {
							return ctrl.Result{Requeue: true}, nil
						} else {
							return ctrl.Result{}, errors.New("failed to update status", err)
						}
					}
					return ctrl.Result{}, errors.New("unsupported application strategy: %s", app.Spec.Strategy.Missing)
				}
			}

			// Find the corresponding Deployment object for the current repository & branch
			var deployment *apiv1.Deployment
			for i, d := range deploymentsToDelete {
				dr := d.Spec.Repository
				if dr.APIVersion == repoRef.APIVersion && dr.Kind == repoRef.Kind {
					if dr.Namespace == repoRef.Namespace && dr.Name == repoRef.Name {
						if d.Spec.Branch == strings.TrimPrefix(ref.Spec.Ref, "refs/heads/") {
							deployment = d
							deploymentsToDelete = append(deploymentsToDelete[:i], deploymentsToDelete[i+1:]...)
							break
						}
					}
				}
			}

			// If no corresponding Deployment was found, we need to create a new one for this repository & branch
			if deployment == nil {
				deployment = &apiv1.Deployment{
					TypeMeta: metav1.TypeMeta{APIVersion: apiv1.GroupVersion.String(), Kind: "Deployment"},
					ObjectMeta: metav1.ObjectMeta{
						Name:      stringsutil.RandomHash(16),
						Namespace: env.Namespace,
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion:         apiv1.GroupVersion.String(),
								Kind:               "ApplicationEnvironment",
								Name:               env.Name,
								UID:                env.UID,
								BlockOwnerDeletion: &[]bool{true}[0],
							},
						},
					},
					Spec: apiv1.DeploymentSpec{
						Repository: corev1.ObjectReference{
							APIVersion: apiv1.GroupVersion.String(),
							Kind:       apiv1.GitHubRepositoryGVK.Kind,
							Name:       repo.Name,
							Namespace:  repo.Namespace,
							UID:        repo.UID,
						},
						Branch: strings.TrimPrefix(ref.Spec.Ref, "refs/heads/"),
					},
				}
				if err := r.Create(ctx, deployment); err != nil {
					env.SetStatusConditionValid(metav1.ConditionUnknown, apiv1.ReasonInternalError, fmt.Sprintf("Failed creating deployment: %s", err.Error()))
					if err := r.Status().Update(ctx, env); err != nil {
						if apierrors.IsConflict(err) {
							return ctrl.Result{Requeue: true}, nil
						} else {
							return ctrl.Result{}, errors.New("failed to update status", err)
						}
					}
					return ctrl.Result{}, errors.New("failed creating deployment", err)
				}
			}

		} else {
			env.SetStatusConditionValid(metav1.ConditionUnknown, apiv1.ReasonConfigError, "Unsupported repository kind: "+repoRef.GroupVersionKind().String())
			if err := r.Status().Update(ctx, env); err != nil {
				if apierrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				} else {
					return ctrl.Result{}, errors.New("failed to update status", err)
				}
			}
			return ctrl.Result{}, nil // no point to requeue, this is a configuration error
		}
	}

	// Delete all deployments that are no longer needed
	for _, deployment := range deploymentsToDelete {
		if err := r.Delete(ctx, deployment); err != nil {
			env.SetStatusConditionValid(metav1.ConditionUnknown, apiv1.ReasonInternalError, fmt.Sprintf("Failed deleting deployment: %s", err.Error()))
			if err := r.Status().Update(ctx, env); err != nil {
				if apierrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				} else {
					return ctrl.Result{}, errors.New("failed to update status", err)
				}
			}
			return ctrl.Result{}, errors.New("failed deleting deployment", err)
		}
	}

	if env.SetStatusConditionValidIfDifferent(metav1.ConditionTrue, apiv1.ReasonValid, "Valid") {
		if err := r.Status().Update(ctx, env); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.ApplicationEnvironment{}).
		Complete(r)
}
