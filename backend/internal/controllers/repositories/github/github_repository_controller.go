package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/google/go-github/v56/github"
	"github.com/secureworks/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"slices"
	"strings"
	"time"
)

var (
	RepositoryFinalizer = "github.repositories.finalizers." + apiv1.GroupVersion.Group
)

type RepositoryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	repo := &apiv1.GitHubRepository{}
	if err := r.Get(ctx, req.NamespacedName, repo); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, errors.New("failed to get '%s'", req.NamespacedName, err)
		}
	}

	// Apply finalization if object is deleted
	if repo.GetDeletionTimestamp() != nil {
		if slices.Contains(repo.GetFinalizers(), RepositoryFinalizer) {
			var finalizers []string
			for _, f := range repo.GetFinalizers() {
				if f != RepositoryFinalizer {
					finalizers = append(finalizers, f)
				}
			}
			repo.SetFinalizers(finalizers)
			if err := r.Update(ctx, repo); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it's missing
	if !slices.Contains(repo.GetFinalizers(), RepositoryFinalizer) {
		repo.SetFinalizers(append(repo.GetFinalizers(), RepositoryFinalizer))
		if err := r.Update(ctx, repo); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Initialize conditions
	if repo.GetStatusConditionCurrent() == nil {
		repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonInitializing, "Initializing")
		if err := r.Status().Update(ctx, repo); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}
	if repo.GetStatusConditionAuthenticatedToGitHub() == nil {
		repo.SetStatusConditionAuthenticatedToGitHub(metav1.ConditionUnknown, apiv1.ReasonInitializing, "Initializing")
		if err := r.Status().Update(ctx, repo); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Break early if we're not ready to sync
	if b, r := k8s.ShouldBreak(ctx, repo, "init"); b {
		log.FromContext(ctx).Info("breaking early", "result", r)
		return r, nil
	}

	// Obtain GitHub client
	ghClient, status, reason, message, err := repo.GetGitHubClient(ctx, r.Client)
	if repo.SetStatusConditionAuthenticatedToGitHubIfDifferent(status, reason, message) {
		if err := r.Status().Update(ctx, repo); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
	}
	if err != nil {
		return ctrl.Result{}, err
	} else if ghClient == nil {
		return ctrl.Result{}, nil
	}

	// Map all currently existing GitHubRepositoryRef objects by their branch name
	branches := map[string]*apiv1.GitHubRepositoryRef{}
	gitHubRefObjects := &apiv1.GitHubRepositoryRefList{}
	if err := k8s.GetOwnedChildrenByIndex(ctx, r.Client, repo, gitHubRefObjects); err != nil {
		repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonInternalError, "Failed listing GitHubRepositoryRef objects: "+err.Error())
		if err := r.Status().Update(ctx, repo); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{}, errors.New("failed listing GitHubRepositoryRef objects", err)
	}
	for _, ghRepRefObj := range gitHubRefObjects.Items {
		ref := ghRepRefObj
		branches[strings.TrimPrefix(ghRepRefObj.Spec.Ref, "refs/heads/")] = &ref
	}

	// Fetch branch names
	branchesListOptions := &github.BranchListOptions{}
	for {
		branchesList, response, err := ghClient.Repositories.ListBranches(ctx, repo.Spec.Owner, repo.Spec.Name, branchesListOptions)
		if err != nil {
			repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonGitHubAPIFailed, "Failed listing branches: "+err.Error())
			if err := r.Status().Update(ctx, repo); err != nil {
				if apierrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				} else {
					return ctrl.Result{}, errors.New("failed to update status", err)
				}
			}
			return ctrl.Result{}, errors.New("failed to list branches", err)
		}
		for _, branch := range branchesList {
			if _, ok := branches[branch.GetName()]; !ok {
				branches[branch.GetName()] = nil
			}
		}
		if response.NextPage == 0 {
			break
		}
		branchesListOptions.Page = response.NextPage
	}

	// Ensure a GitHubRepositoryRef object exists for each branch
	for branchName, refObj := range branches {
		if refObj == nil {
			refObject := &apiv1.GitHubRepositoryRef{
				ObjectMeta: metav1.ObjectMeta{
					Name:      stringsutil.RandomHash(7),
					Namespace: repo.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         repo.APIVersion,
							Kind:               repo.Kind,
							Name:               repo.Name,
							UID:                repo.UID,
							BlockOwnerDeletion: &[]bool{true}[0],
						},
					},
				},
				Spec: apiv1.GitHubRepositoryRefSpec{Ref: "refs/heads/" + branchName},
			}
			if err := r.Create(ctx, refObject); err != nil {
				repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonInternalError, "Failed creating GitHubRepositoryRef: "+err.Error())
				if err := r.Status().Update(ctx, repo); err != nil {
					if apierrors.IsConflict(err) {
						return ctrl.Result{Requeue: true}, nil
					} else {
						return ctrl.Result{}, errors.New("failed to update status", err)
					}
				}
				return ctrl.Result{}, errors.New("failed creating GitHubRepositoryRef object", err)
			}
		}
	}

	// Ensure "Current" condition is set to "True"
	if repo.SetStatusConditionCurrentIfDifferent(metav1.ConditionTrue, apiv1.ReasonSynced, "Synchronized refs") {
		if err := r.Status().Update(ctx, repo); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// TODO: make GitHubRepository refresh interval configurable
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.GitHubRepository{}).
		Complete(r)
}
