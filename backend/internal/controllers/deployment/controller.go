package deployment

import (
	"context"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/controllers/deployment/applier"
	"github.com/arikkfir/devbot/backend/internal/controllers/deployment/baker"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/secureworks/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	Finalizer = "deployments.finalizers." + apiv1.GroupVersion.Group
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	baker  *baker.Baker
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.executeReconciliation(ctx, req).ToResultAndError()
}

func (r *Reconciler) executeReconciliation(ctx context.Context, req ctrl.Request) *k8s.Result {
	rec, result := k8s.NewReconciliation(ctx, r.Client, req, &apiv1.Deployment{}, Finalizer, nil)
	if result != nil {
		return result
	}

	// Finalize object if deleted
	if result := rec.FinalizeObjectIfDeleted(); result != nil {
		return result
	}

	// Initialize object if not initialized
	if result := rec.InitializeObject(); result != nil {
		return result
	}

	// Fetch source repository
	repoKey := rec.Object.Spec.Repository.GetObjectKey()
	repo := &apiv1.Repository{}
	if err := r.Client.Get(rec.Ctx, repoKey, repo); err != nil {
		if apierrors.IsNotFound(err) {
			rec.Object.Status.SetMaybeStaleDueToRepositoryNotFound("Repository '%s' could not be found", repoKey)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		} else if apierrors.IsForbidden(err) {
			rec.Object.Status.SetMaybeStaleDueToRepositoryNotAccessible("Repository '%s' is not accessible: %+v", repoKey, err)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		} else {
			rec.Object.Status.SetMaybeStaleDueToInternalError("Failed looking up repository '%s': %+v", repoKey, err)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		}
	}

	// Get controlling environment
	env := &apiv1.Environment{}
	if result := rec.GetRequiredController(env); result != nil {
		return result
	}

	// Get controlling application
	var app *apiv1.Application
	if result := r.getApplication(rec, env, &app); result != nil {
		return result
	}

	// Infer clone path and store it in the status object
	if rec.Object.Status.ClonePath == "" {
		rec.Object.Status.SetStaleDueToCloneMissing("Clone path not set yet")
		if result := rec.UpdateStatus(); result != nil {
			return result
		}

		rec.Object.Status.ClonePath = fmt.Sprintf("/data/%s/%s/%s/%s/%s", app.Namespace, app.Name, env.Name, repo.Namespace, repo.Name)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
	}

	// Make sure the parent directory of the target clone path exists
	parentDir := filepath.Dir(rec.Object.Status.ClonePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		rec.Object.Status.SetMaybeStaleDueToCloneMissing("Failed creating directory '%s': %+v", parentDir, err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	} else {
		rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.CloneMissing)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
	}

	// Calculate Git URL from repository
	if _, err := os.Stat(rec.Object.Status.ClonePath); errors.Is(err, os.ErrNotExist) {

		// Clone path does not exist - calculate Git URL based on repository type, and clone it
		var url string
		if repo.Spec.GitHub != nil {
			url = fmt.Sprintf("https://github.com/%s/%s", repo.Spec.GitHub.Owner, repo.Spec.GitHub.Name)
		} else {
			rec.Object.Status.SetInvalidDueToRepositoryNotSupported("Unsupported repository")
			rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.DoNotRequeue()
		}

		// Set cloning status
		rec.Object.Status.SetMaybeStaleDueToCloning("Cloning repository")
		if result := rec.UpdateStatus(); result != nil {
			return result
		}

		// Clone
		cloneOptions := &git.CloneOptions{
			URL:      url,
			Progress: os.Stdout, // TODO: represent progress in status object
		}
		if _, err := git.PlainClone(rec.Object.Status.ClonePath, false, cloneOptions); err != nil {
			rec.Object.Status.SetMaybeStaleDueToCloneFailed("Failed cloning repository: %+v", err)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			if err := os.RemoveAll(rec.Object.Status.ClonePath); err != nil {
				return k8s.RequeueDueToError(errors.New("failed removing clone directory after failed clone: %w", err))
			} else {
				return k8s.Requeue()
			}
		}

	} else if err != nil {
		rec.Object.Status.SetStaleDueToInternalError("Failed inspecting target clone directory '%s': %+v", rec.Object.Status.ClonePath, err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Open the cloned repository
	gitRepo, err := git.PlainOpen(rec.Object.Status.ClonePath)
	if err != nil {
		rec.Object.Status.SetMaybeStaleDueToCloneOpenFailed("Failed opening cloned repository: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Infer the branch to deploy
	var sha, branch string
	if revision, ok := repo.Status.Revisions[env.Spec.PreferredBranch]; ok {
		branch = env.Spec.PreferredBranch
		sha = revision
	} else if revision, ok := repo.Status.Revisions[repo.Status.DefaultBranch]; ok {
		branch = repo.Status.DefaultBranch
		sha = revision
	} else {
		rec.Object.Status.SetMaybeStaleDueToBranchNotFound("Neither branch '%s' nor '%s' found in repository '%s'", env.Spec.PreferredBranch, repo.Status.DefaultBranch, client.ObjectKeyFromObject(repo))
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Update last-attempted revision
	rec.Object.Status.LastAttemptedRevision = sha
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Fetch our branch
	localBranchRefName := plumbing.NewBranchReferenceName(branch)
	remoteBranchRefName := plumbing.NewRemoteReferenceName("origin", branch)
	refSpec := fmt.Sprintf("%s:%s", localBranchRefName, remoteBranchRefName)
	fetchOptions := git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec(refSpec)},
		Progress:   os.Stdout,
	}
	if err := gitRepo.FetchContext(rec.Ctx, &fetchOptions); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		rec.Object.Status.SetMaybeStaleDueToFetchFailed("Failed fetching branch '%s': %+v", branch, err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Attempt to open the worktree
	worktree, err := gitRepo.Worktree()
	if err != nil {
		rec.Object.Status.SetMaybeStaleDueToCloneOpenFailed("Failed opening repository worktree: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Checkout the exact revision listed in the repository
	if err := worktree.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(sha)}); err != nil {
		rec.Object.Status.SetMaybeStaleDueToCheckoutFailed("Failed checking out revision '%s' of branch '%s': %+v", sha, branch, err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Signal we're baking
	rec.Object.Status.SetMaybeStaleDueToBaking("Baking resources")
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Bake resources manifest
	b := baker.NewDefaultBaker()
	manifestFile, err := b.GenerateManifest(rec.Ctx, rec.Object.Status.ClonePath, app.Name, env.Spec.PreferredBranch, branch, sha)
	if err != nil {
		rec.Object.Status.SetMaybeStaleDueToBakingFailed("Failed baking resources: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	} else if resourcesManifest, err := os.ReadFile(manifestFile); err != nil {
		rec.Object.Status.SetMaybeStaleDueToBakingFailed("Failed reading generated resources manifest: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	} else {
		rec.Object.Status.ResourcesManifest = string(resourcesManifest)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
	}

	// Signal we're applying manifest
	rec.Object.Status.SetMaybeStaleDueToApplying("Applying manifest")
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Apply resources
	a := applier.NewDefaultApplier()
	if err := a.Apply(rec.Ctx, manifestFile); err != nil {
		rec.Object.Status.SetMaybeStaleDueToApplyFailed("Failed applying resources manifest: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Update last-applied revision
	rec.Object.Status.SetCurrent()
	rec.Object.Status.LastAppliedRevision = sha
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	return k8s.DoNotRequeue()
}

func (r *Reconciler) getApplication(rec *k8s.Reconciliation[*apiv1.Deployment], env *apiv1.Environment, appTarget **apiv1.Application) *k8s.Result {
	appRef := metav1.GetControllerOf(env)
	if appRef == nil {
		rec.Object.Status.SetInvalidDueToInternalError("Could not find application controller for parent environment '%s/%s'", env.Namespace, env.Name)
		rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.DoNotRequeue()
	}

	app := &apiv1.Application{}
	if err := r.Client.Get(rec.Ctx, client.ObjectKey{Name: appRef.Name, Namespace: env.Namespace}, app); err != nil {
		if apierrors.IsNotFound(err) {
			rec.Object.Status.SetInvalidDueToControllerNotFound("Could not find application controller of parent environment: %+v", err)
			rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		} else if apierrors.IsForbidden(err) {
			rec.Object.Status.SetInvalidDueToControllerNotAccessible("Application controller of parent environment is not accessible: %+v", err)
			rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		} else {
			rec.Object.Status.SetInvalidDueToInternalError("Failed to get application controller of parent environment: %+v", err)
			rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		}
	}
	rec.Object.Status.SetValidIfInvalidDueToAnyOf(apiv1.ControllerNotAccessible, apiv1.ControllerNotFound, apiv1.InternalError)
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.Invalid)
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	*appTarget = app
	return k8s.Continue()
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.baker = baker.NewDefaultBaker()
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Deployment{}).
		Complete(r)
}

/*func deployment := o.(*apiv1.Deployment)
		resourcesContent, err := os.ReadFile(resourcesFile)
		if err != nil {
			deployment.Status.SetMaybeStaleDueToInternalError("Failed reading resources file: %+v", err)
			return reconcile.Requeue()
		}

		var resources []unstructured.Unstructured
		yamlDecoder := yaml.NewDecoder(bytes.NewReader(resourcesContent))
		for {
			var yamlDocument yaml.Node
			if err := yamlDecoder.Decode(&yamlDocument); err != nil {
				if err == io.EOF {
					break
				}
				deployment.Status.SetMaybeStaleDueToInternalError("Failed decoding resources YAML: %+v", err)
				return reconcile.Requeue()
			}

			var yamlBytes bytes.Buffer
			if err := yaml.NewEncoder(&yamlBytes).Encode(yamlDocument); err != nil {
				deployment.Status.SetMaybeStaleDueToInternalError("Failed encoding back to YAML string: %+v", err)
				return reconcile.Requeue()
			}

			jsonData, err := yamlutil.ToJSON(yamlBytes.Bytes())
			if err != nil {
				deployment.Status.SetMaybeStaleDueToInternalError("Failed to convert YAML back to JSON: %+v", err)
				return reconcile.Requeue()
			}

			unstructuredObject := &unstructured.Unstructured{}
			if _, _, err = unstructured.UnstructuredJSONScheme.Decode(jsonData, nil, unstructuredObject); err != nil {
				deployment.Status.SetMaybeStaleDueToInternalError("Failed decoding JSON to unstructured object: %+v", err)
				return reconcile.Requeue()
			}

			gvk := unstructuredObject.GetObjectKind().GroupVersionKind()
			gvr, _ := meta.UnsafeGuessKindToResource(gvk)
			name := unstructuredObject.GetName()
			namespace := unstructuredObject.GetNamespace()
			if namespace == "" {
				deployment.Status.SetMaybeStaleDueToInvalid("Resource '%s' has no namespace", name)
				return reconcile.Requeue()
			}
			resourceClient := dynamicClient.Resource(gvr).Namespace(namespace)

			if _, err = resourceClient.Get(ctx, name, metav1.GetOptions{}); err != nil {
				if apierrors.IsNotFound(err) {
					if _, err := resourceClient.Create(ctx, unstructuredObject, metav1.CreateOptions{}); err != nil {
						deployment.Status.SetMaybeStaleDueToInternalError("Failed to create: %+v", err)
						return reconcile.Requeue()
					} else {
						resources = append(resources, *unstructuredObject)
					}
				}
			} else {
				_, err := dynamicClient.Resource(gvr).Namespace(ns).Update(context.Background(), unstructuredObj, metav1.UpdateOptions{})
				if err != nil {
					log.Fatalf("Failed to update: %v", err)
				}
			}
}
*/
