package application

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewCreateMissingEnvironmentsAction(envList *apiv1.EnvironmentList) reconcile.Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		app := o.(*apiv1.Application)

		// Create a map of all environments by their preferred branch
		envByBranch := make(map[string]*apiv1.Environment)
		for i, item := range envList.Items {
			envByBranch[item.Spec.PreferredBranch] = &envList.Items[i]
		}

		// Ensure all branches from all participating repositories are mapped to an environment
		for _, repoRef := range app.Spec.Repositories {
			repoClientKey := repoRef.GetObjectKey(app.Namespace)
			if repoRef.IsGitHubRepository() {

				// This is a GitHub repository
				repo := &apiv1.GitHubRepository{}
				if err := c.Get(ctx, repoClientKey, repo); err != nil {
					if apierrors.IsNotFound(err) {
						app.Status.SetStaleDueToRepositoryNotFound("GitHub repository '%s' not found", repoClientKey)
						return reconcile.Requeue()
					} else if apierrors.IsForbidden(err) {
						app.Status.SetMaybeStaleDueToRepositoryNotAccessible("GitHub repository '%s' is not accessible: %+v", repoClientKey, err)
						return reconcile.Requeue()
					} else {
						app.Status.SetMaybeStaleDueToInternalError("Failed looking up GitHub repository '%s': %+v", repoClientKey, err)
						return reconcile.Requeue()
					}
				}

				// Get all GitHubRepositoryRef objects that belong to this repository
				refs := &apiv1.GitHubRepositoryRefList{}
				if err := c.List(ctx, refs, k8s.OwnedBy(c.Scheme(), repo)); err != nil {
					app.Status.SetMaybeStaleDueToInternalError("Failed looking up refs for GitHub repository '%s': %+v", repoClientKey, err)
					return reconcile.Requeue()
				}

				// For every repository branch found, ensure there's an environment for it
				for _, ref := range refs.Items {
					if _, ok := envByBranch[ref.Spec.Ref]; !ok {
						env := &apiv1.Environment{
							ObjectMeta: metav1.ObjectMeta{
								Name:      strings.RandomHash(7),
								Namespace: app.Namespace,
								OwnerReferences: []metav1.OwnerReference{
									*metav1.NewControllerRef(app, apiv1.ApplicationGVK),
								},
							},
							Spec: apiv1.EnvironmentSpec{PreferredBranch: ref.Spec.Ref},
						}
						if err := c.Create(ctx, env); err != nil {
							app.Status.SetMaybeStaleDueToInternalError("Failed creating environment for branch '%s': %+v", ref.Spec.Ref, err)
							return reconcile.Requeue()
						}
					}
				}

			} else {
				app.Status.SetInvalidDueToRepositoryNotSupported("Unsupported repository type '%s.%s' specified", repoRef.Kind, repoRef.APIVersion)
				app.Status.SetMaybeStaleDueToInvalid(app.Status.GetInvalidMessage())
				return reconcile.DoNotRequeue()
			}
		}
		return reconcile.Continue()
	}
}
