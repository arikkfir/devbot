package deployment

import (
	"context"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewPrepareCloneDirectoryAction(targetDir *string, gitURL *string) reconcile.Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		d := o.(*apiv1.Deployment)

		// Get the repository
		var repo client.Object
		var url string
		if d.Spec.Repository.IsGitHubRepository() {
			repo = &apiv1.GitHubRepository{}
		} else {
			d.Status.SetInvalidDueToRepositoryNotSupported("Unsupported repository type '%s.%s' specified", d.Spec.Repository.Kind, d.Spec.Repository.APIVersion)
			d.Status.SetMaybeStaleDueToInvalid(d.Status.GetInvalidMessage())
			return reconcile.DoNotRequeue()
		}
		repoKey := d.Spec.Repository.GetObjectKey()
		if err := c.Get(ctx, repoKey, repo); err != nil {
			if apierrors.IsNotFound(err) {
				d.Status.SetInvalidDueToRepositoryNotFound("Repository '%s' not found", repoKey)
				d.Status.SetMaybeStaleDueToInvalid(d.Status.GetInvalidMessage())
				return reconcile.Requeue()
			} else if apierrors.IsForbidden(err) {
				d.Status.SetInvalidDueToRepositoryNotAccessible("Repository '%s' is not accessible: %+v", repoKey, err)
				d.Status.SetMaybeStaleDueToInvalid(d.Status.GetInvalidMessage())
				return reconcile.Requeue()
			} else {
				d.Status.SetMaybeStaleDueToInternalError("Failed looking up repository '%s': %+v", repoKey, err)
				return reconcile.Requeue()
			}
		} else {
			switch r := repo.(type) {
			case *apiv1.GitHubRepository:
				url = fmt.Sprintf("https://github.com/%s/%s", r.Spec.Owner, r.Spec.Name)
			default:
				d.Status.SetInvalidDueToRepositoryNotSupported("Unsupported repository type '%s.%s' specified", d.Spec.Repository.Kind, d.Spec.Repository.APIVersion)
				return reconcile.DoNotRequeue()
			}
		}

		// Decide on a clone path and save it in the object
		if d.Status.ClonePath == "" {
			if dir, err := os.MkdirTemp(os.TempDir(), "clone-*"); err != nil {
				d.Status.SetStaleDueToCloneFailed("Failed creating temporary directory: %+v", err)
				return reconcile.Requeue()
			} else if err := os.RemoveAll(dir); err != nil {
				d.Status.SetStaleDueToCloneFailed("Failed removing temporary directory: %+v", err)
				return reconcile.Requeue()
			} else {
				d.Status.SetStaleDueToCloneMissing("Repository not cloned yet")
				d.Status.ClonePath = dir
				return reconcile.Requeue()
			}
		}

		*targetDir = d.Status.ClonePath
		*gitURL = url
		d.Status.SetMaybeStaleDueToCloning("Opening/cloning repository from '%s' into '%s'", gitURL, d.Status.ClonePath)
		return reconcile.Continue()
	}
}
