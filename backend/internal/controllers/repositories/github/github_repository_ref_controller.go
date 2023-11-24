package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util"
	"github.com/google/go-github/v56/github"
	"github.com/secureworks/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type RepositoryRefReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *RepositoryRefReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ref := &apiv1.GitHubRepositoryRef{}

	// Ensure deletion & finalizers lifecycle processing
	if result, err := util.PrepareReconciliation(ctx, r.Client, req, ref, "refs.github.repositories.finalizers."+apiv1.GroupVersion.Group); result != nil || err != nil {
		return *result, err
	}

	// Initialize "Current" condition
	if ref.GetStatusConditionCurrent() == nil {
		ref.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonInitializing, "Initializing")
		if err := r.Status().Update(ctx, ref); err != nil {
			return ctrl.Result{}, errors.New("failed to update status", err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Find own owning GitHubRepository
	repo := &apiv1.GitHubRepository{}
	if err := util.GetOwnerOfObject(ctx, r.Client, ref, apiv1.GroupVersion.String(), "GitHubRepository", repo); err != nil {
		ref.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonOwnerMissing, "Owner not found: "+err.Error())
		if err := r.Status().Update(ctx, ref); err != nil {
			return ctrl.Result{}, errors.New("failed to update status", err)
		}
		return ctrl.Result{}, err
	}

	// Obtain GitHub authentication
	personalAccessToken, err := getPersonalAccessToken(ctx, r.Client, repo)
	if err != nil {
		ref.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonConfigError, "Cannot get personal access token: "+err.Error())
		if err := r.Status().Update(ctx, ref); err != nil {
			return ctrl.Result{}, errors.New("failed to update status", err)
		}
		return ctrl.Result{}, errors.New("failed to get personal access token", err)
	}
	ghClient := github.NewClient(nil).WithAuthToken(personalAccessToken)

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
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
			return ctrl.Result{}, errors.New("failed to query branch", err)

		}

	}

	// If our branch's commit SHA is different than the one we have in our status, update it
	if branch.Commit.GetSHA() != ref.Status.CommitSHA {
		ref.Status.CommitSHA = branch.Commit.GetSHA()
		if err := r.Status().Update(ctx, ref); err != nil {
			return ctrl.Result{}, errors.New("failed to update status", err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// If our branch's commit SHA is the same as the one we have in our status, set condition "Current" to "True"
	if ref.SetStatusConditionCurrentIfDifferent(metav1.ConditionTrue, apiv1.ReasonCurrent, "Commit SHA up-to-date") {
		if err := r.Status().Update(ctx, ref); err != nil {
			return ctrl.Result{}, errors.New("failed to update status", err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepositoryRefReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.GitHubRepositoryRef{}).
		Complete(r)
}
