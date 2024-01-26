package environment

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewSyncSourcesAction(app *apiv1.Application) reconcile.Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		env := o.(*apiv1.Environment)

		// Remove sources that are no longer participating in the application, and update sources that are using the
		// wrong branch (i.e. as branches get created or deleted, this affects which branch should be deployed)
		for srcIndex, envSrc := range env.Status.Sources {
			participating := false
			strategy := apiv1.MissingBranchStrategyIgnore
			for _, appRepo := range app.Spec.Repositories {
				nsAppRepo := appRepo.ToNamespacedRepositoryReference(app.Namespace)
				if nsAppRepo == envSrc.Repository {
					participating = true
					strategy = appRepo.MissingBranchStrategy
					break
				}
			}

			// Fetch deployment of this source
			var deployment *apiv1.Deployment
			if envSrc.Deployment != nil {
				deployment = &apiv1.Deployment{}
				dKey := client.ObjectKey{Namespace: env.Namespace, Name: deployment.Name}
				if err := c.Get(ctx, dKey, deployment); err != nil {
					if apierrors.IsNotFound(err) {
						envSrc.Deployment = nil
						env.Status.SetInvalidDueToDeploymentNotFound("Deployment '%s' is missing", dKey)
						env.Status.SetMaybeStaleDueToInvalid(env.Status.GetInvalidMessage())
						return reconcile.Requeue()
					} else {
						env.Status.SetInvalidDueToInternalError("Failed to get deployment '%s': %+v", dKey, err)
						env.Status.SetMaybeStaleDueToInvalid(env.Status.GetInvalidMessage())
						return reconcile.Requeue()
					}
				}
			}

			// Remove source if its repository is no longer participating in the application
			if !participating {
				if deployment != nil {
					if err := c.Delete(ctx, deployment); err != nil {
						key := client.ObjectKeyFromObject(deployment)
						if apierrors.IsNotFound(err) {
							// Ignore
						} else if apierrors.IsConflict(err) {
							return reconcile.Requeue()
						} else {
							env.Status.SetMaybeStaleDueToInternalError("Failed to delete deployment '%s': %+v", key, err)
							return reconcile.Requeue()
						}
					}
				}
				env.Status.Sources = append(env.Status.Sources[:srcIndex], env.Status.Sources[srcIndex+1:]...)
				return reconcile.Requeue()
			}

			// Ensure source branch is using the correct branch
			repoClientKey := envSrc.Repository.GetObjectKey()
			if envSrc.Repository.IsGitHubRepository() {

				// This is a GitHub repository
				repo := &apiv1.GitHubRepository{}
				if err := c.Get(ctx, repoClientKey, repo); err != nil {
					if apierrors.IsNotFound(err) {
						env.Status.SetStaleDueToRepositoryNotFound("GitHub repository '%s' not found", repoClientKey)
						return reconcile.Requeue()
					} else if apierrors.IsForbidden(err) {
						env.Status.SetMaybeStaleDueToRepositoryNotAccessible("GitHub repository '%s' is not accessible: %+v", repoClientKey, err)
						return reconcile.Requeue()
					} else {
						env.Status.SetMaybeStaleDueToInternalError("Failed looking up GitHub repository '%s': %+v", repoClientKey, err)
						return reconcile.Requeue()
					}
				}

				// Get all GitHubRepositoryRef objects that belong to this repository
				refs := &apiv1.GitHubRepositoryRefList{}
				if err := c.List(ctx, refs, k8s.OwnedBy(c.Scheme(), repo)); err != nil {
					env.Status.SetMaybeStaleDueToInternalError("Failed looking up refs for GitHub repository '%s': %+v", repoClientKey, err)
					return reconcile.Requeue()
				}

				// If branch exists, ensure this source uses that branch and not the default branch
				// Otherwise, use the default branch if that's the strategy for this repository
				// Otherwise, delete this source
				preferredBranchFound := false
				for _, ref := range refs.Items {
					if ref.Spec.Ref == env.Spec.PreferredBranch {
						preferredBranchFound = true
						break
					}
				}

				// Determine the correct branch to deploy; set to empty string if no this source should not be deployed
				var branch string
				if preferredBranchFound {
					branch = env.Spec.PreferredBranch
				} else if strategy == apiv1.MissingBranchStrategyUseDefaultBranch {
					if repo.Status.DefaultBranch == "" {
						env.Status.SetMaybeStaleDueToRepositoryMissingDefaultBranch("GitHub repository '%s' has no default branch", repoClientKey)
						return reconcile.Requeue()
					} else {
						branch = repo.Status.DefaultBranch
					}
				} else if strategy == apiv1.MissingBranchStrategyIgnore {
					branch = ""
				}

				// Apply it to the deployment's specification:
				// - if deployment is missing, create it
				// - if branch is empty - i.e. no branch should be deployed - delete the deployment
				if branch == "" {
					if deployment != nil {
						if err := c.Delete(ctx, deployment); err != nil {
							key := client.ObjectKeyFromObject(deployment)
							if apierrors.IsConflict(err) {
								env.Status.SetMaybeStaleDueToInvalid("Failed to delete deployment '%s': %+v", key, err)
								return reconcile.Requeue()
							} else {
								env.Status.SetInvalidDueToInternalError("Failed to delete deployment '%s': %+v", key, err)
								env.Status.SetMaybeStaleDueToInvalid(env.Status.GetInvalidMessage())
								return reconcile.Requeue()
							}
						}
					}
					envSrc.Deployment = nil
				} else if deployment == nil {
					deployment = &apiv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:            envSrc.Deployment.Name,
							Namespace:       env.Namespace,
							OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(env, apiv1.EnvironmentGVK)},
						},
						Spec: apiv1.DeploymentSpec{
							Repository: envSrc.Repository,
							Branch:     branch,
						},
					}
					if err := c.Create(ctx, deployment); err != nil {
						if !apierrors.IsAlreadyExists(err) {
							env.Status.SetMaybeStaleDueToInternalError("Failed to create deployment '%s': %+v", deployment.Name, err)
							return reconcile.Requeue()
						}
					}
					envSrc.Deployment = &apiv1.DeploymentReference{Name: deployment.Name}
				} else if deployment.Spec.Branch != branch {
					deployment.Spec.Branch = branch
					if err := c.Update(ctx, deployment); err != nil {
						key := client.ObjectKeyFromObject(deployment)
						if apierrors.IsConflict(err) {
							env.Status.SetMaybeStaleDueToInvalid("Failed to update deployment '%s': %+v", key, err)
							return reconcile.Requeue()
						} else {
							env.Status.SetInvalidDueToInternalError("Failed to update deployment '%s': %+v", key, err)
							env.Status.SetMaybeStaleDueToInvalid(env.Status.GetInvalidMessage())
							return reconcile.Requeue()
						}
					}
				}

			} else {
				env.Status.SetInvalidDueToRepositoryNotSupported("Unsupported repository type '%s.%s' specified", envSrc.Repository.Kind, envSrc.Repository.APIVersion)
				return reconcile.DoNotRequeue()
			}

		}

		for _, appRepo := range app.Spec.Repositories {
			nsAppRepo := appRepo.ToNamespacedRepositoryReference(app.Namespace)
			found := false
			for _, envSrc := range env.Status.Sources {
				if nsAppRepo == envSrc.Repository {
					found = true
					break
				}
			}
			if !found {
				env.Status.SetMaybeStaleDueToRepositoryAdded("Missing repository source added: %s/%s", nsAppRepo.Namespace, nsAppRepo.Name)
				env.Status.Sources = append(env.Status.Sources, apiv1.EnvironmentStatusSource{Repository: nsAppRepo})
			}
		}

		return reconcile.Continue()
	}
}
