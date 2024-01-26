package deployment

import (
	"bytes"
	"context"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	"os/exec"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewApplyAction(resourcesFile string) reconcile.Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		deployment := o.(*apiv1.Deployment)

		// TODO: support remote clusters

		stdout := bytes.Buffer{}
		stderr := bytes.Buffer{}
		kubectlCmd := exec.CommandContext(ctx,
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
			deployment.Status.SetMaybeStaleDueToApplyFailed("Failed applying resources: %+v", err)
			return reconcile.Requeue()
		}

		// TODO: infer inventory list from kubectl output (for potential pruning/health checks)

		return reconcile.Continue()
	}
}

/*func NewApplyAction(app *apiv1.Application, env *apiv1.Environment, ghRepo *git.Repository, dynamicClient *dynamic.DynamicClient, resourcesFile string) reconcile.Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		deployment := o.(*apiv1.Deployment)


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

		reference, err := ghRepo.Head()
		if err != nil {
			deployment.Status.SetMaybeStaleDueToInternalError("Failed getting git head: %+v", err)
			return reconcile.Requeue()
		}
		commitSHA := reference.Hash().String()
		if deployment.Status.LastAppliedCommitSHA == commitSHA {
			// Nothing to do!
			return reconcile.DoNotRequeue()
		}

		root := deployment.Status.ClonePath
		possibleKustomizationFilePaths := []string{
			filepath.Join(root, ".devbot", app.Name, strings.Slugify(env.Spec.PreferredBranch), "kustomization.yaml"),
			filepath.Join(root, ".devbot", app.Name, "kustomization.yaml"),
			filepath.Join(root, ".devbot", strings.Slugify(env.Spec.PreferredBranch), "kustomization.yaml"),
			filepath.Join(root, ".devbot", "kustomization.yaml"),
		}

		var kustomizationFilePath string
		for _, path := range possibleKustomizationFilePaths {
			if info, err := os.Stat(path); err != nil {
				if !os.IsNotExist(err) {
					deployment.Status.SetMaybeStaleDueToInternalError("Failed checking if kustomization file exists at '%s': %+v", path, err)
					return reconcile.Requeue()
				}
			} else if info.IsDir() {
				deployment.Status.SetMaybeStaleDueToInternalError("Kustomization file at '%s' is a directory", path)
				return reconcile.Requeue()
			} else {
				kustomizationFilePath = path
				break
			}
		}
		if kustomizationFilePath == "" {
			deployment.Status.SetMaybeStaleDueToInternalError("Failed locating kustomization file in '%s'", root)
			return reconcile.Requeue()
		}

		// Create target resources file
		resourcesFile, err := os.Create(filepath.Dir(kustomizationFilePath) + "/.devbot.output.resources.yaml")
		if err != nil {
			deployment.Status.SetMaybeStaleDueToInternalError("Failed creating resources file: %+v", err)
			return reconcile.Requeue()
		}
		defer resourcesFile.Close()

		// Create a pipe that connects stdout of the "kustomize build" command to the "kustomize fn" command
		pipeReader, pipeWriter := io.Pipe()

		// This command produces resources from the kustomization file and outputs them to stdout
		kustomizeBuildCmd := exec.CommandContext(ctx,
			kustomizeBinaryFilePath,
			"build",
			filepath.Dir(kustomizationFilePath),
		)
		kustomizeBuildCmd.Dir = filepath.Dir(kustomizationFilePath)
		kustomizeBuildCmd.Stderr = &bytes.Buffer{}
		kustomizeBuildCmd.Stdout = pipeWriter

		// This command accepts resources via stdin, processes them via the bash function script, and outputs to stdout
		kustomizeFnCmd := exec.CommandContext(ctx, yqBinaryFilePath, `(.. | select(tag == "!!str")) |= envsubst`)
		kustomizeFnCmd.Env = append(os.Environ(),
			"APPLICATION="+app.Name,
			"BRANCH="+deployment.Spec.Branch,
			"COMMIT_SHA="+commitSHA,
			"ENVIRONMENT="+env.Spec.PreferredBranch,
		)
		kustomizeBuildCmd.Dir = filepath.Dir(kustomizationFilePath)
		kustomizeFnCmd.Stderr = &bytes.Buffer{}
		kustomizeFnCmd.Stdin = pipeReader
		kustomizeFnCmd.Stdout = resourcesFile

		// Execute kustomize build
		if err := kustomizeBuildCmd.Start(); err != nil {
			deployment.Status.SetMaybeStaleDueToInternalError("Failed starting kustomize command: %+v", err)
			return reconcile.Requeue()
		}
		defer kustomizeBuildCmd.Process.Kill()

		// Execute kustomize fn
		if err := kustomizeFnCmd.Start(); err != nil {
			deployment.Status.SetMaybeStaleDueToInternalError("Failed starting kustomize fn: %+v", err)
			return reconcile.Requeue()
		}
		defer kustomizeFnCmd.Process.Kill()

		// Wait for both commands to finish
		if err := kustomizeBuildCmd.Wait(); err != nil {
			deployment.Status.SetMaybeStaleDueToInternalError("Failed executing kustomize build: %+v", err)
			return reconcile.Requeue()
		} else if err := kustomizeFnCmd.Wait(); err != nil {
			deployment.Status.SetMaybeStaleDueToInternalError("Failed executing kustomize fn: %+v", err)
			return reconcile.Requeue()
		}

		// If kustomize failed, set condition and exit
		if kustomizeBuildCmd.ProcessState.ExitCode()+kustomizeFnCmd.ProcessState.ExitCode() > 0 {
			stderr := bytes.Buffer{}
			stderr.WriteString("--[kustomize build stderr:]--------------------------------------------------\n")
			stderr.Write(kustomizeBuildCmd.Stderr.(*bytes.Buffer).Bytes())
			stderr.WriteString("\n--[kustomize fn stderr:]--------------------------------------------------\n")
			stderr.Write(kustomizeFnCmd.Stderr.(*bytes.Buffer).Bytes())
			deployment.Status.SetStaleDueToKustomizeFailure("Kustomize failed:\n%s", stderr.String())
			return reconcile.Requeue()
		}

		*targetResourcesFile = resourcesFile.Name()
		return reconcile.Continue()
	}
}
*/
