package deployment

import (
	"bytes"
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	"github.com/go-git/go-git/v5"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewBakeAction(app *apiv1.Application, env *apiv1.Environment, ghRepo *git.Repository, targetResourcesFile *string) reconcile.Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		deployment := o.(*apiv1.Deployment)

		reference, err := ghRepo.Head()
		if err != nil {
			deployment.Status.SetMaybeStaleDueToBakingFailed("Failed getting git head: %+v", err)
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
					deployment.Status.SetMaybeStaleDueToBakingFailed("Failed checking if kustomization file exists at '%s': %+v", path, err)
					return reconcile.Requeue()
				}
			} else if info.IsDir() {
				deployment.Status.SetMaybeStaleDueToBakingFailed("Kustomization file at '%s' is a directory", path)
				return reconcile.Requeue()
			} else {
				kustomizationFilePath = path
				break
			}
		}
		if kustomizationFilePath == "" {
			deployment.Status.SetMaybeStaleDueToBakingFailed("Failed locating kustomization file in '%s'", root)
			return reconcile.Requeue()
		}

		// Create target resources file
		resourcesFile, err := os.Create(filepath.Dir(kustomizationFilePath) + "/.devbot.output.resources.yaml")
		if err != nil {
			deployment.Status.SetMaybeStaleDueToBakingFailed("Failed creating resources file: %+v", err)
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
		yqCmd := exec.CommandContext(ctx, yqBinaryFilePath, `(.. | select(tag == "!!str")) |= envsubst`)
		yqCmd.Env = append(os.Environ(),
			"APPLICATION="+app.Name,
			"BRANCH="+deployment.Spec.Branch,
			"COMMIT_SHA="+commitSHA,
			"ENVIRONMENT="+env.Spec.PreferredBranch,
		)
		yqCmd.Dir = filepath.Dir(kustomizationFilePath)
		yqCmd.Stderr = &bytes.Buffer{}
		yqCmd.Stdin = pipeReader
		yqCmd.Stdout = resourcesFile

		// Execute kustomize build
		if err := kustomizeBuildCmd.Start(); err != nil {
			deployment.Status.SetMaybeStaleDueToBakingFailed("Failed starting kustomize command: %+v", err)
			return reconcile.Requeue()
		}
		defer kustomizeBuildCmd.Process.Kill()

		// Execute kustomize fn
		if err := yqCmd.Start(); err != nil {
			deployment.Status.SetMaybeStaleDueToBakingFailed("Failed starting kustomize fn: %+v", err)
			return reconcile.Requeue()
		}
		defer yqCmd.Process.Kill()

		// Wait for both commands to finish
		if err := kustomizeBuildCmd.Wait(); err != nil {
			deployment.Status.SetMaybeStaleDueToBakingFailed("Failed executing kustomize build: %+v", err)
			return reconcile.Requeue()
		} else if err := yqCmd.Wait(); err != nil {
			deployment.Status.SetMaybeStaleDueToBakingFailed("Failed executing kustomize fn: %+v", err)
			return reconcile.Requeue()
		}

		// If kustomize failed, set condition and exit
		if kustomizeBuildCmd.ProcessState.ExitCode()+yqCmd.ProcessState.ExitCode() > 0 {
			stderr := bytes.Buffer{}
			stderr.WriteString("--[kustomize build stderr:]--------------------------------------------------\n")
			stderr.Write(kustomizeBuildCmd.Stderr.(*bytes.Buffer).Bytes())
			stderr.WriteString("\n--[yq stderr:]---------------------------------------------------------------\n")
			stderr.Write(yqCmd.Stderr.(*bytes.Buffer).Bytes())
			deployment.Status.SetStaleDueToBakingFailed("Manifest baking failed:\n%s", stderr.String())
			return reconcile.Requeue()
		}

		*targetResourcesFile = resourcesFile.Name()
		return reconcile.Continue()
	}
}
