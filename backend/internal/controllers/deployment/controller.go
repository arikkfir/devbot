package deployment

import (
	"bytes"
	"context"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/secureworks/errors"
	"io"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"os/exec"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	kustomizeBinaryFilePath = "/usr/local/bin/kustomize"
	yqBinaryFilePath        = "/usr/local/bin/yq"
	kubectlBinaryFilePath   = "/usr/local/bin/kubectl"
)

var (
	Finalizer = "deployments.finalizers." + apiv1.GroupVersion.Group
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rec, result := k8s.NewReconciliation(ctx, r.Client, req, &apiv1.Deployment{}, Finalizer, nil)
	if result != nil {
		return result.ToResultAndError()
	}

	// Finalize object if deleted
	if result := rec.FinalizeObjectIfDeleted(); result != nil {
		return result.ToResultAndError()
	}

	// Initialize object if not initialized
	if result := rec.InitializeObject(); result != nil {
		return result.ToResultAndError()
	}

	// Get controlling environment
	env := &apiv1.Environment{}
	if result := rec.GetRequiredController(env); result != nil {
		return result.ToResultAndError()
	}

	// Get controlling application
	var app *apiv1.Application
	if result := r.getApplicationController(rec, env, &app); result != nil {
		return result.ToResultAndError()
	}

	// Clone repository
	var gitURL string
	var ghRepo *git.Repository
	if result := r.clone(rec, &gitURL, &ghRepo); result != nil {
		return result.ToResultAndError()
	}

	// Bake resources manifest
	var commitSHA, resourcesFile string
	if result := r.bake(rec, app, env, ghRepo, &commitSHA, &resourcesFile); result != nil {
		return result.ToResultAndError()
	}

	// Apply resources manifest
	if result := r.apply(rec, resourcesFile); result != nil {
		return result.ToResultAndError()
	}

	// Update last-applied commit SHA
	rec.Object.Status.LastAppliedCommitSHA = commitSHA
	if result := rec.UpdateStatus(); result != nil {
		return result.ToResultAndError()
	}

	return k8s.DoNotRequeue().ToResultAndError()
}

// TODO: review all Requeue results and consider replacing with DoNotRequeue (thus rely on next polling event)
//       also review error states - on many cases we probably should re-clone the repository

func (r *Reconciler) getApplicationController(rec *k8s.Reconciliation[*apiv1.Deployment], env *apiv1.Environment, appTarget **apiv1.Application) *k8s.Result {
	appControllerRef := metav1.GetControllerOf(env)
	if appControllerRef == nil {
		rec.Object.Status.SetInvalidDueToInternalError("Could not find application controller reference in parent environment")
		rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.DoNotRequeue()
	}

	app := &apiv1.Application{}
	if err := r.Client.Get(rec.Ctx, client.ObjectKey{Name: appControllerRef.Name, Namespace: env.Namespace}, app); err != nil {
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

func (r *Reconciler) clone(rec *k8s.Reconciliation[*apiv1.Deployment], gitURL *string, ghRepo **git.Repository) *k8s.Result {
	// Get the repository
	var url string
	if rec.Object.Spec.Repository.IsGitHubRepository() {
		repo := &apiv1.GitHubRepository{}
		repoKey := rec.Object.Spec.Repository.GetObjectKey()
		if err := r.Get(rec.Ctx, repoKey, repo); err != nil {
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
		url = fmt.Sprintf("https://github.com/%s/%s", repo.Spec.Owner, repo.Spec.Name)
		rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.RepositoryNotFound, apiv1.RepositoryNotAccessible, apiv1.InternalError)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
	} else {
		rec.Object.Status.SetInvalidDueToRepositoryNotSupported("Unsupported repository type '%s.%s' specified", rec.Object.Spec.Repository.Kind, rec.Object.Spec.Repository.APIVersion)
		rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.DoNotRequeue()
	}

	// Decide on a clone path and save it in the object
	if rec.Object.Status.ClonePath == "" {
		rec.Object.Status.SetStaleDueToCloneMissing("Clone path not set yet")
		if result := rec.UpdateStatus(); result != nil {
			return result
		}

		rec.Object.Status.ClonePath = fmt.Sprintf("/data/%s/%s/%s", rec.Object.Spec.Repository.Namespace, rec.Object.Spec.Repository.Name, stringsutil.RandomHash(7))
		if result := rec.UpdateStatus(); result != nil {
			return result
		}

		rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.CloneMissing)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
	}

	// Inspect the clone path; if it does not exist, clone the repository
	if _, err := os.Stat(rec.Object.Status.ClonePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {

			rec.Object.Status.SetMaybeStaleDueToCloning("Cloning repository")
			if result := rec.UpdateStatus(); result != nil {
				return result
			}

			cloneOptions := &git.CloneOptions{
				URL:      url,
				Progress: os.Stdout, // TODO: represent progress in status object
				Depth:    1,
			}
			if _, err := git.PlainClone(rec.Object.Status.ClonePath, false, cloneOptions); err != nil {
				rec.Object.Status.SetMaybeStaleDueToCloneFailed("Failed cloning repository: %+v", err)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
				return k8s.Requeue()
			}

		} else {
			// TODO: consider auto-healing by deleting the whole directory
			rec.Object.Status.SetMaybeStaleDueToInternalError("Failed to stat local clone dir at '%s': %+v", rec.Object.Status.ClonePath, err)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		}
	}

	// Attempt to open the cloned repository
	repo, err := git.PlainOpen(rec.Object.Status.ClonePath)
	if err != nil {
		// TODO: consider auto-healing by deleting the whole directory
		rec.Object.Status.SetMaybeStaleDueToCloneFailed("Failed opening cloned repository: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Attempt to open the worktree
	worktree, err := repo.Worktree()
	if err != nil {
		// TODO: consider auto-healing by deleting the whole directory
		rec.Object.Status.SetMaybeStaleDueToCloneFailed("Failed opening repository worktree: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Set status to pulling
	rec.Object.Status.SetMaybeStaleDueToPulling("Pulling updates")
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Ensure we're on the correct branch
	refName := plumbing.NewBranchReferenceName(rec.Object.Spec.Branch)
	if err := worktree.Checkout(&git.CheckoutOptions{Branch: refName}); err != nil {
		// TODO: consider auto-healing by deleting the whole directory
		rec.Object.Status.SetMaybeStaleDueToCheckoutFailed("Failed checking out branch '%s': %+v", rec.Object.Spec.Branch, err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Ensure we're up-to-date
	if err := worktree.PullContext(rec.Ctx, &git.PullOptions{SingleBranch: true, ReferenceName: refName}); err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			// TODO: consider auto-healing by deleting the whole directory
			rec.Object.Status.SetMaybeStaleDueToPullFailed("Failed pulling changes: %+v", err)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		}
	}

	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.RepositoryNotFound, apiv1.RepositoryNotAccessible, apiv1.InternalError, apiv1.CloneMissing, apiv1.Cloning, apiv1.CloneFailed, apiv1.CheckoutFailed, apiv1.Pulling, apiv1.PullFailed)
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	*gitURL = url
	*ghRepo = repo
	return k8s.Continue()
}

func (r *Reconciler) bake(rec *k8s.Reconciliation[*apiv1.Deployment], app *apiv1.Application, env *apiv1.Environment, ghRepo *git.Repository, commitSHA, targetResourcesFile *string) *k8s.Result {
	reference, err := ghRepo.Head()
	if err != nil {
		rec.Object.Status.SetMaybeStaleDueToBakingFailed("Failed getting Git HEAD revision: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Find the kustomization
	root := rec.Object.Status.ClonePath
	possibleKustomizationFilePaths := []string{
		filepath.Join(root, ".devbot", app.Name, stringsutil.Slugify(env.Spec.PreferredBranch), "kustomization.yaml"),
		filepath.Join(root, ".devbot", app.Name, stringsutil.Slugify(rec.Object.Spec.Branch), "kustomization.yaml"),
		filepath.Join(root, ".devbot", stringsutil.Slugify(rec.Object.Spec.Branch), "kustomization.yaml"),
		filepath.Join(root, ".devbot", stringsutil.Slugify(env.Spec.PreferredBranch), "kustomization.yaml"),
	}
	var kustomizationFilePath string
	for _, path := range possibleKustomizationFilePaths {
		if info, err := os.Stat(path); err != nil {
			if !os.IsNotExist(err) {
				rec.Object.Status.SetMaybeStaleDueToBakingFailed("Failed checking if kustomization file exists at '%s': %+v", path, err)
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
				return k8s.Requeue()
			}
		} else if info.IsDir() {
			// TODO: for current SHA we've basically failed; we should not requeue until new SHA is available
			rec.Object.Status.SetMaybeStaleDueToBakingFailed("Kustomization file at '%s' is a directory", path)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		} else {
			kustomizationFilePath = path
			break
		}
	}
	if kustomizationFilePath == "" {
		// TODO: for current SHA we've basically failed; we should not requeue until new SHA is available
		rec.Object.Status.SetMaybeStaleDueToBakingFailed("Failed locating kustomization file in '%s'", root)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Signal we're building manifests
	rec.Object.Status.SetMaybeStaleDueToBakingFailed("Building resources manifest from kustomization at: %s", kustomizationFilePath)
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Create target resources file
	resourcesFile, err := os.Create(filepath.Dir(kustomizationFilePath) + "/.devbot.output.resources.yaml")
	if err != nil {
		rec.Object.Status.SetMaybeStaleDueToBakingFailed("Failed creating resources file: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}
	defer resourcesFile.Close()

	// Create a pipe that connects stdout of the "kustomize build" command to the "kustomize fn" command
	pipeReader, pipeWriter := io.Pipe()

	// This command produces resources from the kustomization file and outputs them to stdout
	kustomizeBuildCmd := exec.CommandContext(rec.Ctx, kustomizeBinaryFilePath, "build", filepath.Dir(kustomizationFilePath))
	kustomizeBuildCmd.Dir = filepath.Dir(kustomizationFilePath)
	kustomizeBuildCmd.Stderr = &bytes.Buffer{}
	kustomizeBuildCmd.Stdout = pipeWriter

	// This command accepts resources via stdin, processes them via the bash function script, and outputs to stdout
	yqCmd := exec.CommandContext(rec.Ctx, yqBinaryFilePath, `(.. | select(tag == "!!str")) |= envsubst`)
	yqCmd.Env = append(os.Environ(),
		"APPLICATION="+stringsutil.Slugify(app.Name),
		"BRANCH="+stringsutil.Slugify(rec.Object.Spec.Branch),
		"COMMIT_SHA="+reference.Hash().String(),
		"ENVIRONMENT="+stringsutil.Slugify(env.Spec.PreferredBranch),
	)
	yqCmd.Dir = filepath.Dir(kustomizationFilePath)
	yqCmd.Stderr = &bytes.Buffer{}
	yqCmd.Stdin = pipeReader
	yqCmd.Stdout = resourcesFile

	// Execute kustomize build
	if err := kustomizeBuildCmd.Start(); err != nil {
		rec.Object.Status.SetMaybeStaleDueToBakingFailed("Failed starting kustomize command: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}
	defer kustomizeBuildCmd.Process.Kill()

	// Execute kustomize fn
	if err := yqCmd.Start(); err != nil {
		rec.Object.Status.SetMaybeStaleDueToBakingFailed("Failed starting kustomize fn: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}
	defer yqCmd.Process.Kill()

	// Wait for both commands to finish
	if err := kustomizeBuildCmd.Wait(); err != nil {
		rec.Object.Status.SetMaybeStaleDueToBakingFailed("Failed executing kustomize build: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	} else if err := yqCmd.Wait(); err != nil {
		rec.Object.Status.SetMaybeStaleDueToBakingFailed("Failed executing kustomize fn: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// If kustomize failed, set condition and exit
	if kustomizeBuildCmd.ProcessState.ExitCode()+yqCmd.ProcessState.ExitCode() > 0 {
		stderr := bytes.Buffer{}
		stderr.WriteString("--[kustomize build stderr:]--------------------------------------------------\n")
		stderr.Write(kustomizeBuildCmd.Stderr.(*bytes.Buffer).Bytes())
		stderr.WriteString("\n--[yq stderr:]---------------------------------------------------------------\n")
		stderr.Write(yqCmd.Stderr.(*bytes.Buffer).Bytes())
		rec.Object.Status.SetStaleDueToBakingFailed("Manifest baking failed:\n%s", stderr.String())
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Signal we're baked
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.BakingFailed)
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	*targetResourcesFile = resourcesFile.Name()
	*commitSHA = reference.Hash().String()
	return k8s.Continue()
}

func (r *Reconciler) apply(rec *k8s.Reconciliation[*apiv1.Deployment], resourcesFile string) *k8s.Result {
	// TODO: support remote clusters

	// Signal we're applying manifest
	rec.Object.Status.SetMaybeStaleDueToApplying("Applying manifest")
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Apply resources
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	kubectlCmd := exec.CommandContext(rec.Ctx,
		kubectlBinaryFilePath,
		"apply",
		fmt.Sprintf("--filename=%s", resourcesFile),
		fmt.Sprintf("--dry-run='%s'", "server"),
		fmt.Sprintf("--server-side=%v", true),
	)
	kubectlCmd.Dir = filepath.Dir(resourcesFile)
	kubectlCmd.Stderr = &stderr
	kubectlCmd.Stdout = &stdout
	if err := kubectlCmd.Run(); err != nil {
		rec.Object.Status.SetMaybeStaleDueToApplyFailed("Failed applying resources: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	log.FromContext(rec.Ctx).Info("kubectl apply", "stdout", stdout.String(), "stderr", stderr.String())

	// TODO: infer inventory list from kubectl output (for potential pruning/health checks)

	return k8s.Continue()
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
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
				_, err := dynamicClient.Resource(gvr).Namespace(ns).Update(context.TODO(), unstructuredObj, metav1.UpdateOptions{})
				if err != nil {
					log.Fatalf("Failed to update: %v", err)
				}
			}
}
*/
