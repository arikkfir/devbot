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

type GitHubRepositoryRefReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *GitHubRepositoryRefReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ref := &apiv1.GitHubRepositoryRef{}

	if result, err := util.PrepareReconciliation(ctx, r.Client, req, ref, "refs.github.repositories.finalizers."+apiv1.GroupVersion.Group); result != nil || err != nil {
		return *result, err
	}

	// Get owning GitHubRepository object
	if len(ref.OwnerReferences) == 0 {
		ref.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonOwnerMissing, "Owner references empty")
		if err := r.Status().Update(ctx, ref); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", ref.Namespace, ref.Name, err)
		}
		return ctrl.Result{}, errors.New("owner references empty")
	}
	var owner *metav1.OwnerReference
	for _, or := range ref.OwnerReferences {
		if or.Kind == "GitHubRepository" {
			owner = &or
			break
		}
	}
	if owner == nil {
		ref.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonOwnerMissing, "Owner reference to GitHubRepository not found")
		if err := r.Status().Update(ctx, ref); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", ref.Namespace, ref.Name, err)
		}
		return ctrl.Result{}, errors.New("owner reference to GitHubRepository not found")
	}
	repo := &apiv1.GitHubRepository{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: owner.Name}, repo); err != nil {
		ref.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonOwnerMissing, "GitHubRepository missing")
		if err := r.Status().Update(ctx, ref); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", ref.Namespace, ref.Name, err)
		}
		return ctrl.Result{}, errors.New("failed to get owner GitHubRepository '%s/%s'", req.Namespace, owner.Name, err)
	}

	// Build GitHub client
	personalAccessToken, err := getPersonalAccessToken(ctx, r.Client, repo)
	if err != nil {
		ref.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonAuthError, "Personal access token missing")
		if err := r.Status().Update(ctx, ref); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", ref.Namespace, ref.Name, err)
		}
		return ctrl.Result{}, errors.New("failed to get personal access token from '%s/%s': %w", req.Namespace, owner.Name, err)
	}
	ghClient := github.NewClient(nil).WithAuthToken(personalAccessToken)

	// Update commit SHA in ref status
	if strings.HasPrefix(ref.Spec.Ref, "refs/heads/") {
		branch, response, err := ghClient.Repositories.GetBranch(ctx, repo.Spec.Owner, repo.Spec.Name, strings.TrimPrefix(ref.Spec.Ref, "refs/heads/"), 0)
		if err != nil {
			ref.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonGitHubAPIFailed, "Failed getting branch: "+err.Error())
			if err := r.Status().Update(ctx, ref); err != nil {
				return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", ref.Namespace, ref.Name, err)
			}
			return ctrl.Result{}, errors.New("failed to get branch '%s': %w", ref.Spec.Ref, err)
		} else if branch == nil {
			if response.StatusCode == http.StatusNotFound {
				if err := r.Delete(ctx, ref); err != nil {
					return ctrl.Result{}, errors.New("failed to delete branch '%s/%s': %w", ref.Namespace, ref.Name, err)
				}
				return ctrl.Result{}, nil
			}
		}
		if branch.Commit.GetSHA() != ref.Status.CommitSHA {
			ref.Status.CommitSHA = branch.Commit.GetSHA()
			ref.SetStatusConditionCurrent(metav1.ConditionTrue, apiv1.ReasonCurrent, "Commit SHA up-to-date")
			if err := r.Status().Update(ctx, ref); err != nil {
				return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", ref.Namespace, ref.Name, err)
			}
		}
	} else if strings.HasPrefix(ref.Spec.Ref, "refs/tags/") {
		listTagsOpts := &github.ListOptions{}
		for {
			tags, _, err := ghClient.Repositories.ListTags(ctx, repo.Spec.Owner, repo.Spec.Name, listTagsOpts)
			if err != nil {
				ref.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonGitHubAPIFailed, "Failed getting tag: "+err.Error())
				if err := r.Status().Update(ctx, ref); err != nil {
					return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", ref.Namespace, ref.Name, err)
				}
				return ctrl.Result{}, errors.New("failed to get tag '%s': %w", ref.Spec.Ref, err)
			}
			found := false
			for _, tag := range tags {
				if tag.GetName() == strings.TrimPrefix(ref.Spec.Ref, "refs/tags/") {
					if tag.Commit.GetSHA() != ref.Status.CommitSHA {
						ref.Status.CommitSHA = tag.Commit.GetSHA()
						ref.SetStatusConditionCurrent(metav1.ConditionTrue, apiv1.ReasonCurrent, "Commit SHA up-to-date")
						if err := r.Status().Update(ctx, ref); err != nil {
							return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", ref.Namespace, ref.Name, err)
						}
					}
					found = true
					break
				}
			}
			if !found {
				if err := r.Delete(ctx, ref); err != nil {
					return ctrl.Result{}, errors.New("failed to delete tag '%s/%s': %w", ref.Namespace, ref.Name, err)
				}
				return ctrl.Result{}, nil
			}
		}
	} else {
		ref.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonInternalError, "Invalid ref: "+ref.Spec.Ref)
		if err := r.Status().Update(ctx, ref); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", ref.Namespace, ref.Name, err)
		}
		return ctrl.Result{}, errors.New("invalid ref '%s'", ref.Spec.Ref)
	}

	ref.SetStatusConditionCurrent(metav1.ConditionTrue, apiv1.ReasonCurrent, "Current")
	if err := r.Status().Update(ctx, ref); err != nil {
		return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", ref.Namespace, ref.Name, err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GitHubRepositoryRefReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.GitHubRepositoryRef{}).
		Complete(r)
}
