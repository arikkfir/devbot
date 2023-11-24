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
	"strings"
)

var (
	RepositoryFinalizer = "github.repositories.finalizers." + apiv1.GroupVersion.Group
)

type RepositoryReconciler struct {
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

func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	repo := &apiv1.GitHubRepository{}

	// Ensure deletion & finalizers lifecycle processing
	if result, err := util.PrepareReconciliation(ctx, r.Client, req, repo, RepositoryFinalizer); result != nil || err != nil {
		return *result, err
	}

	// Initialize "Current" condition
	if repo.GetStatusConditionCurrent() == nil {
		repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonInitializing, "Initializing")
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status", err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Obtain GitHub authentication
	personalAccessToken, err := getPersonalAccessToken(ctx, r.Client, repo)
	if err != nil {
		repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonConfigError, "Cannot get personal access token: "+err.Error())
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status", err)
		}
		return ctrl.Result{}, errors.New("failed to get personal access token", err)
	}
	ghClient := github.NewClient(nil).WithAuthToken(personalAccessToken)

	// Fetch branch names
	branches := map[string]*apiv1.GitHubRepositoryRef{}
	branchesListOptions := &github.BranchListOptions{}
	for {
		branchesList, response, err := ghClient.Repositories.ListBranches(ctx, repo.Spec.Owner, repo.Spec.Name, branchesListOptions)
		if err != nil {
			repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonGitHubAPIFailed, "Failed listing branches: "+err.Error())
			if err := r.Status().Update(ctx, repo); err != nil {
				return ctrl.Result{}, errors.New("failed to update status", err)
			}
			return ctrl.Result{}, errors.New("failed to list branches", err)
		}
		for _, branch := range branchesList {
			branches[branch.GetName()] = nil
		}
		if response.NextPage == 0 {
			break
		}
		branchesListOptions.Page = response.NextPage
	}

	// Fetch all GitHubRepositoryRef objects owned by this GitHubRepository
	gitHubRefObjects := &apiv1.GitHubRepositoryRefList{}
	if err := r.List(ctx, gitHubRefObjects, client.InNamespace(repo.Namespace), client.MatchingFields{util.OwnerRefField: util.GetOwnerRefKey(repo)}); err != nil {
		repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonInternalError, "Failed listing GitHubRepositoryRef objects: "+err.Error())
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status", err)
		}
		return ctrl.Result{}, errors.New("failed listing GitHubRepositoryRef objects", err)
	}
	for _, ghRepRefObj := range gitHubRefObjects.Items {
		branches[strings.TrimPrefix(ghRepRefObj.Spec.Ref, "refs/heads/")] = &ghRepRefObj
	}

	// Ensure a GitHubRepositoryRef object exists for each branch
	for branchName, refObj := range branches {
		if refObj == nil {
			refObject := &apiv1.GitHubRepositoryRef{
				ObjectMeta: metav1.ObjectMeta{
					Name:      util.RandomHash(7),
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
					return ctrl.Result{}, errors.New("failed to update status", err)
				}
				return ctrl.Result{}, errors.New("failed creating GitHubRepositoryRef object", err)
			}
		}
	}

	// Ensure "Current" condition is set to "True"
	if repo.SetStatusConditionCurrentIfDifferent(metav1.ConditionTrue, "Synced", "Synchronized refs") {
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status", err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()

	if err := mgr.GetFieldIndexer().IndexField(ctx, &apiv1.GitHubRepositoryRef{}, "metadata.ownerRef", util.IndexableOwnerReferences); err != nil {
		return errors.New("failed to index field 'metadata.ownerRef' of 'GitHubRepositoryRef' objects", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.GitHubRepository{}).
		Complete(r)
}
