package deployment

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/secureworks/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewCloneRepositoryAction(url string, ghRepo **git.Repository) reconcile.Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		d := o.(*apiv1.Deployment)

		// Set status to cloning
		d.Status.SetMaybeStaleDueToCloning("Cloning repository")
		if err := c.Status().Update(ctx, d); err != nil {
			if apierrors.IsNotFound(err) {
				return reconcile.DoNotRequeue()
			} else if apierrors.IsConflict(err) {
				return reconcile.Requeue()
			} else {
				return reconcile.RequeueDueToError(err)
			}
		}

		// Attempt to open the clone directory
		if r, err := git.PlainOpen(d.Status.ClonePath); err != nil {
			if errors.Is(err, git.ErrRepositoryNotExists) {
				// Clone
				_, err := git.PlainClone(d.Status.ClonePath, false, &git.CloneOptions{
					URL:      url,
					Progress: os.Stdout,
					Depth:    1,
				})
				if err != nil {
					d.Status.SetMaybeStaleDueToCloneFailed("Failed cloning repository: %+v", err)
					return reconcile.Requeue()
				}
				return reconcile.Requeue()
			} else {
				d.Status.SetMaybeStaleDueToCloneFailed("Failed opening repository: %+v", err)
				return reconcile.Requeue()
			}
		} else if worktree, err := r.Worktree(); err != nil {
			d.Status.SetMaybeStaleDueToCloneFailed("Failed opening repository worktree: %+v", err)
			return reconcile.Requeue()
		} else {

			// Set status to pulling
			d.Status.SetMaybeStaleDueToPulling("Pulling updates")
			if err := c.Status().Update(ctx, d); err != nil {
				if apierrors.IsNotFound(err) {
					return reconcile.DoNotRequeue()
				} else if apierrors.IsConflict(err) {
					return reconcile.Requeue()
				} else {
					return reconcile.RequeueDueToError(err)
				}
			}

			refName := plumbing.NewBranchReferenceName(d.Spec.Branch)
			if err := worktree.Checkout(&git.CheckoutOptions{Branch: refName}); err != nil {
				d.Status.SetMaybeStaleDueToCloneFailed("Failed checking out branch '%s': %+v", d.Spec.Branch, err)
				return reconcile.Requeue()
			}

			err := worktree.PullContext(ctx, &git.PullOptions{
				SingleBranch:  true,
				ReferenceName: refName,
			})
			if err != nil {
				if errors.Is(err, git.NoErrAlreadyUpToDate) {
					// No changes
				} else {
					d.Status.SetMaybeStaleDueToCloneFailed("Failed pulling changes: %+v", err)
					return reconcile.Requeue()
				}
			}

			d.Status.SetCurrent()
			*ghRepo = r
			return reconcile.Continue()
		}
	}
}
