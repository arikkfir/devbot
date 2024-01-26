package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	"github.com/google/go-github/v56/github"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func NewFetchGitHubRepositoryAction(refreshInterval time.Duration, gh *github.Client, ghRepo **github.Repository) reconcile.Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		r := o.(*apiv1.GitHubRepository)
		if r.Spec.Owner == "" {
			r.Status.SetInvalidDueToRepositoryOwnerMissing("Repository owner is empty")
			r.Status.SetMaybeStaleDueToInvalid(r.Status.GetInvalidMessage())
			return reconcile.DoNotRequeue()
		} else if r.Spec.Name == "" {
			r.Status.SetInvalidDueToRepositoryNameMissing("Repository name is empty")
			r.Status.SetMaybeStaleDueToInvalid(r.Status.GetInvalidMessage())
			return reconcile.DoNotRequeue()
		} else if ghr, resp, err := gh.Repositories.Get(ctx, r.Spec.Owner, r.Spec.Name); err != nil {
			if resp.StatusCode == http.StatusNotFound {
				r.Status.SetMaybeStaleDueToRepositoryNotFound("Repository not found: %s", resp.Status)
				return reconcile.RequeueAfter(refreshInterval)
			} else {
				r.Status.SetMaybeStaleDueToGitHubAPIFailed("Failed fetching repository '%s/%s': %+v", r.Spec.Owner, r.Spec.Name, err)
				return reconcile.RequeueAfter(refreshInterval)
			}
		} else {
			*ghRepo = ghr
			return reconcile.Continue()
		}
	}
}
