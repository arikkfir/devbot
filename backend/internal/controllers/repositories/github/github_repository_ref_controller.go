package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/secureworks/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"slices"
	"strings"
	"time"
)

var (
	RepositoryRefFinalizer = "github.repositories.refs.finalizers." + apiv1.GroupVersion.Group
)

type RepositoryRefReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *RepositoryRefReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ref := &apiv1.GitHubRepositoryRef{}
	if err := r.Get(ctx, req.NamespacedName, ref); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, errors.New("failed to get '%s'", req.NamespacedName, err)
		}
	}

	// Apply finalization if object is deleted
	if ref.GetDeletionTimestamp() != nil {
		if slices.Contains(ref.GetFinalizers(), RepositoryRefFinalizer) {
			var finalizers []string
			for _, f := range ref.GetFinalizers() {
				if f != RepositoryRefFinalizer {
					finalizers = append(finalizers, f)
				}
			}
			ref.SetFinalizers(finalizers)
			if err := r.Update(ctx, ref); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it's missing
	if !slices.Contains(ref.GetFinalizers(), RepositoryRefFinalizer) {
		ref.SetFinalizers(append(ref.GetFinalizers(), RepositoryRefFinalizer))
		if err := r.Update(ctx, ref); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Initialize conditions
	if ref.GetStatusConditionCurrent() == nil {
		ref.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonInitializing, "Initializing")
		if err := r.Status().Update(ctx, ref); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}
	if ref.GetStatusConditionAuthenticatedToGitHub() == nil {
		ref.SetStatusConditionAuthenticatedToGitHub(metav1.ConditionUnknown, apiv1.ReasonInitializing, "Initializing")
		if err := r.Status().Update(ctx, ref); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Break early if we're not ready to sync
	if b, r := k8s.ShouldBreak(ctx, ref, "init"); b {
		log.FromContext(ctx).Info("breaking early", "result", r)
		return r, nil
	}

	// Find own owning GitHubRepository
	repo := &apiv1.GitHubRepository{}
	if err := k8s.GetFirstOwnerOfType(ctx, r.Client, ref, apiv1.GroupVersion.String(), apiv1.GitHubRepositoryGVK.Kind, repo); err != nil {
		ref.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonInternalError, "Owner not found: "+err.Error())
		if err := r.Status().Update(ctx, ref); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{}, nil
	}

	// Obtain GitHub client
	ghClient, status, reason, message, err := repo.GetGitHubClient(ctx, r.Client)
	if ref.SetStatusConditionAuthenticatedToGitHubIfDifferent(status, reason, message) {
		if err := r.Status().Update(ctx, ref); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	// Ensure our ref still exists in GitHub
	branch, resp, err := ghClient.Repositories.GetBranch(ctx, repo.Spec.Owner, repo.Spec.Name, strings.TrimPrefix(ref.Spec.Ref, "refs/heads/"), 0)
	if err != nil {

		if resp.StatusCode == http.StatusNotFound {

			// Our branch no longer exists in GitHub, delete this object from Kubernetes
			if err := r.Delete(ctx, ref); err != nil {
				return ctrl.Result{}, errors.New("failed to delete object", err)
			}
			return ctrl.Result{}, nil

		} else {

			// GitHub API failed us for some reason
			ref.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonGitHubAPIFailed, "Failed querying branch: "+err.Error())
			if err := r.Status().Update(ctx, ref); err != nil {
				if apierrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				} else {
					return ctrl.Result{}, errors.New("failed to update status", err)
				}
			}
			return ctrl.Result{}, errors.New("failed to query branch", err)

		}

	}

	// If our branch's commit SHA is different from the one we have in our status, update it
	if branch.Commit.GetSHA() != ref.Status.CommitSHA {
		ref.Status.CommitSHA = branch.Commit.GetSHA()
		if err := r.Status().Update(ctx, ref); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// If our branch's commit SHA is the same as the one we have in our status, set condition "Current" to "True"
	if ref.SetStatusConditionCurrentIfDifferent(metav1.ConditionTrue, apiv1.ReasonSynced, "Commit SHA up-to-date") {
		if err := r.Status().Update(ctx, ref); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepositoryRefReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.GitHubRepositoryRef{}).
		Watches(&apiv1.GitHubRepository{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			repo := obj.(*apiv1.GitHubRepository)
			refs := &apiv1.GitHubRepositoryRefList{}
			if err := k8s.GetOwnedChildrenByIndex(ctx, mgr.GetClient(), repo, refs); err != nil {
				log.FromContext(ctx).Error(err, "Failed to get owned children", "owner", repo.Namespace+"/"+repo.Name)
				return nil
			}
			var requests []reconcile.Request
			for _, ref := range refs.Items {
				requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKey{Namespace: ref.Namespace, Name: ref.Name}})
			}
			return requests
		})).
		Complete(r)
}
