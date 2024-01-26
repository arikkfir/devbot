package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	"github.com/google/go-github/v56/github"
	"github.com/secureworks/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewCreateMissingGitHubRepositoryRefObjectsAction(branches []*github.Branch, refsList *apiv1.GitHubRepositoryRefList) reconcile.Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		r := o.(*apiv1.GitHubRepository)
		for _, branch := range branches {
			found := false
			for _, ref := range refsList.Items {
				if ref.Spec.Ref == branch.GetName() {
					found = true
					break
				}
			}
			if !found {
				r.Status.SetStaleDueToBranchesOutOfSync("Branch without GitHubRepositoryRef object found: %s", branch.GetName())
				ref := &apiv1.GitHubRepositoryRef{
					ObjectMeta: metav1.ObjectMeta{
						Name:            stringsutil.RandomHash(7),
						Namespace:       r.Namespace,
						OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(r, apiv1.GitHubRepositoryGVK)},
					},
					Spec: apiv1.GitHubRepositoryRefSpec{Ref: branch.GetName()},
				}
				if err := c.Create(ctx, ref); err != nil {
					return reconcile.RequeueDueToError(errors.New("failed creating GitHubRepositoryRef object for branch '%s': %w", branch.GetName(), err))
				} else {
					return reconcile.Requeue()
				}
			}
		}
		return reconcile.Continue()
	}
}
