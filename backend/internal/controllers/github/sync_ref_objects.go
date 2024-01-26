package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	"github.com/google/go-github/v56/github"
	"github.com/secureworks/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewSyncGitHubRepositoryRefObjectsAction(branches []*github.Branch, refsList *apiv1.GitHubRepositoryRefList) reconcile.Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		r := o.(*apiv1.GitHubRepository)
		for _, ref := range refsList.Items {
			var branch *github.Branch
			var commitSHA string
			for _, b := range branches {
				if b.GetName() == ref.Spec.Ref {
					branch = b
					if b.Commit != nil {
						commitSHA = b.Commit.GetSHA()
					}
					break
				}
			}

			if branch == nil {
				r.Status.SetStaleDueToBranchesOutOfSync("Branches out of sync")
				if err := c.Delete(ctx, &ref); err != nil {
					return reconcile.RequeueDueToError(errors.New("failed deleting GitHubRepositoryRef object '%s/%s': %w", ref.Namespace, ref.Name, err))
				} else {
					return reconcile.Requeue()
				}
			}

			refChanged := false
			if r.Spec.Owner != ref.Status.RepositoryOwner {
				ref.Status.SetStaleDueToRepositoryOwnerOutOfSync("Repository owner '%s' is stale (expected '%s')", ref.Status.RepositoryOwner, r.Spec.Owner)
				ref.Status.RepositoryOwner = r.Spec.Owner
				refChanged = true
			}
			if r.Spec.Name != ref.Status.RepositoryName {
				ref.Status.SetStaleDueToRepositoryNameOutOfSync("Repository name '%s' is stale (expected '%s')", ref.Status.RepositoryName, r.Spec.Name)
				ref.Status.RepositoryName = r.Spec.Name
				refChanged = true
			}
			if commitSHA != ref.Status.CommitSHA {
				ref.Status.SetStaleDueToCommitSHAOutOfSync("Commit SHA '%s' is stale (expected '%s')", ref.Status.CommitSHA, branch.Commit.GetSHA())
				ref.Status.CommitSHA = commitSHA
				refChanged = true
			}

			if refChanged {
				if err := c.Status().Update(ctx, &ref); err != nil {
					return reconcile.RequeueDueToError(errors.New("failed updating GitHubRepositoryRef object '%s/%s': %w", ref.Namespace, ref.Name, err))
				} else {
					return reconcile.Requeue()
				}
			}
		}
		return reconcile.Continue()
	}
}
