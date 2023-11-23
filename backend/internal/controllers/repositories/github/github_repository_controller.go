package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util"
	"github.com/google/go-github/v56/github"
	"github.com/secureworks/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	GitHubRepositoryFinalizer = "github.repositories.finalizers." + apiv1.GroupVersion.Group
)

type GitHubRepositoryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type ConditionInstruction struct {
	True    bool
	False   bool
	Unknown bool
	Reason  string
	Message string
}

func (r *GitHubRepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	repo := &apiv1.GitHubRepository{}

	// Ensure deletion & finalizers lifecycle processing
	if result, err := util.PrepareReconciliation(ctx, r.Client, req, repo, GitHubRepositoryFinalizer); result != nil || err != nil {
		return *result, err
	}

	// Build authentication
	personalAccessToken, err := getPersonalAccessToken(ctx, r.Client, repo)
	if err != nil {
		repo.SetStatusConditionCurrent(metav1.ConditionFalse, apiv1.ReasonAuthError, "Auth configuration error: "+err.Error())
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
		}
		return ctrl.Result{}, err
	}
	ghClient := github.NewClient(nil).WithAuthToken(personalAccessToken)

	// Fetch branches
	branchesListOptions := &github.BranchListOptions{}
	for {
		branches, response, err := ghClient.Repositories.ListBranches(ctx, repo.Spec.Owner, repo.Spec.Name, branchesListOptions)
		if err != nil {
			return ctrl.Result{}, errors.New("failed to list branches of '%s/%s'", repo.Spec.Owner, repo.Spec.Name, err)
		}
		for _, branch := range branches {
			refName := "refs/heads/" + branch.GetName()
			if err := r.ensureGitHubRepositoryRef(ctx, repo, refName); err != nil {
				repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonRefSyncError, err.Error())
				if err := r.Status().Update(ctx, repo); err != nil {
					return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
				}
				return ctrl.Result{}, errors.New("failed to create ref object for branch '%s'", refName, err)
			}
		}
		if response.NextPage == 0 {
			break
		}
		branchesListOptions.Page = response.NextPage
	}

	// Fetch tags
	tagListOptions := &github.ListOptions{}
	for {
		tags, response, err := ghClient.Repositories.ListTags(ctx, repo.Spec.Owner, repo.Spec.Name, tagListOptions)
		if err != nil {
			return ctrl.Result{}, errors.New("failed to list tags of '%s/%s'", repo.Spec.Owner, repo.Spec.Name, err)
		}
		for _, tag := range tags {
			refName := "refs/tags/" + tag.GetName()
			if err := r.ensureGitHubRepositoryRef(ctx, repo, refName); err != nil {
				repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonRefSyncError, err.Error())
				if err := r.Status().Update(ctx, repo); err != nil {
					return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
				}
				return ctrl.Result{}, errors.New("failed to create ref object for tag '%s'", refName, err)
			}
		}
		if response.NextPage == 0 {
			break
		}
		tagListOptions.Page = response.NextPage
	}

	// Ensure "Current" condition is set to "True"
	if repo.SetStatusConditionCurrentIfDifferent(metav1.ConditionTrue, "Synced", "Synchronized refs") {
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *GitHubRepositoryReconciler) ensureGitHubRepositoryRef(ctx context.Context, repo *apiv1.GitHubRepository, refName string) error {
	refLabels := getGitHubRepositoryRefLabels(repo.Spec.Owner, repo.Spec.Name, refName)
	refObjects := &apiv1.GitHubRepositoryRefList{}
	if err := r.List(ctx, refObjects, client.InNamespace(repo.Namespace), refLabels); err != nil {
		return errors.New("failed listing GitHub repository refs", err)
	}
	if len(refObjects.Items) == 0 {
		refObject := &apiv1.GitHubRepositoryRef{
			ObjectMeta: metav1.ObjectMeta{
				Labels:    refLabels,
				Name:      util.RandomHash(7),
				Namespace: repo.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: repo.APIVersion,
						Kind:       repo.Kind,
						Name:       repo.Name,
						UID:        repo.UID,
					},
				},
			},
			Spec: apiv1.GitHubRepositoryRefSpec{
				Ref: refName,
			},
		}
		if err := r.Create(ctx, refObject); err != nil {
			return errors.New("failed creating ref object '%s/%s' for repo '%s'", refObject.Namespace, refObject.Name, repo.Name, err)
		}

	} else if len(refObjects.Items) > 1 {
		var names []string
		for _, refObj := range refObjects.Items {
			names = append(names, refObj.Name)
		}
		return errors.New("multiple ref objects match repo '%s' and ref '%s': %v", repo.Name, refName, names)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GitHubRepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.GitHubRepository{}).
		Complete(r)
}
