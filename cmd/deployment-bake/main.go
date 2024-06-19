package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/arikkfir/command"
	"github.com/rs/zerolog/log"

	"github.com/arikkfir/devbot/internal/util/lang"
	"github.com/arikkfir/devbot/internal/util/observability"
	stringsutil "github.com/arikkfir/devbot/internal/util/strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// kustomizeBinaryFilePath is the path to the kustomize binary.
	kustomizeBinaryFilePath = "/usr/local/bin/kustomize"

	// yqBinaryFilePath is the path to the yq binary.
	yqBinaryFilePath = "/usr/local/bin/yq"
)

type Action struct {
	ActualBranch      string `required:"true" desc:"Git branch serving as a default, in case the preferred branch was missing."`
	ApplicationName   string `required:"true" desc:"Kubernetes Application object name."`
	BaseDeployDir     string `required:"true" desc:"Base directory Directory holding the Kustomize overlay to build."`
	EnvironmentName   string `required:"true" desc:"Kubernetes Environment object name."`
	DeploymentName    string `required:"true" desc:"Kubernetes Deployment object name."`
	ManifestFile      string `required:"true" desc:"Target file to write resources YAML manifest to."`
	PreferredBranch   string `required:"true" desc:"Git branch preferred for baking, if it exists."`
	RepoDefaultBranch string `required:"true" desc:"The default branch of the repository being deployed."`
	SHA               string `required:"true" desc:"Commit SHA to checkout."`
}

func (e *Action) Run(ctx context.Context) error {
	log.Logger = log.With().
		Str("actualBranch", e.ActualBranch).
		Str("appName", e.ApplicationName).
		Str("envName", e.EnvironmentName).
		Str("deploymentName", e.DeploymentName).
		Str("baseDeployDir", e.BaseDeployDir).
		Str("manifestFile", e.ManifestFile).
		Str("preferredBranch", e.PreferredBranch).
		Str("sha", e.SHA).
		Logger()

	// Create target resources file
	resourcesFile, err := os.Create(e.ManifestFile)
	if err != nil {
		return fmt.Errorf("failed creating target manifest file: %w", err)
	}
	defer resourcesFile.Close()

	// Create a pipe that connects stdout of the "kustomize build" command to the "kustomize fn" command
	pipeReader, pipeWriter := io.Pipe()

	// This command produces resources from the kustomization file and outputs them to stdout
	var kustomizeCmd *exec.Cmd
	searchPaths := lang.Uniq([]string{
		filepath.Join(e.BaseDeployDir, e.PreferredBranch),
		filepath.Join(e.BaseDeployDir, e.ActualBranch),
		filepath.Join(e.BaseDeployDir, e.RepoDefaultBranch),
		e.BaseDeployDir,
	})
	for _, path := range searchPaths {
		log.Info().Str("path", path).Msg("Checking for Kustomize in path")
		if _, err := os.Stat(path); err == nil {
			kustomizeCmd = exec.CommandContext(ctx, kustomizeBinaryFilePath, "build")
			kustomizeCmd.Dir = path
			break
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed inspecting path repository devbot path: %w", err)
		}
	}
	if kustomizeCmd == nil {
		return fmt.Errorf("failed finding Kustomize in any of: %v", searchPaths)
	}
	kustomizeLogger := log.With().
		Str("command", kustomizeBinaryFilePath).
		Str("dir", kustomizeCmd.Dir).
		Strs("env", kustomizeCmd.Env).
		Strs("args", kustomizeCmd.Args).
		Str("output", "stderr").
		Logger()
	kustomizeCmd.Stderr = kustomizeLogger
	kustomizeCmd.Stdout = pipeWriter
	if err := kustomizeCmd.Start(); err != nil {
		return fmt.Errorf("failed starting kustomize command: %w", err)
	}

	// This command accepts resources via stdin, processes them via the bash function script, and outputs to stdout
	yqCmd := exec.CommandContext(ctx, yqBinaryFilePath, `(.. | select(tag == "!!str")) |= envsubst`)
	yqCmd.Dir = kustomizeCmd.Dir
	yqCmd.Env = append(os.Environ(),
		"ACTUAL_BRANCH="+stringsutil.Slugify(e.ActualBranch),
		"APPLICATION="+stringsutil.Slugify(e.ApplicationName),
		"COMMIT_SHA="+e.SHA,
		"ENVIRONMENT="+stringsutil.Slugify(e.PreferredBranch),
		"PREFERRED_BRANCH="+stringsutil.Slugify(e.PreferredBranch),
	)
	yqLogger := log.With().
		Str("command", yqBinaryFilePath).
		Str("dir", yqCmd.Dir).
		Strs("env", yqCmd.Env).
		Strs("args", yqCmd.Args).
		Logger()
	yqStderrLogger := yqLogger.With().Str("output", "stderr").Logger()
	yqStdoutLogger := yqLogger.With().Str("output", "stdout").Logger()
	yqCmd.Stdin = pipeReader
	yqCmd.Stderr = yqStderrLogger
	yqCmd.Stdout = io.MultiWriter(resourcesFile, yqStdoutLogger)
	if err := yqCmd.Start(); err != nil {
		return fmt.Errorf("failed starting yq command: %w", err)
	}

	// Wait for kustomize command to finish
	if err := kustomizeCmd.Wait(); err != nil {
		return fmt.Errorf("failed running kustomize command: %w", err)
	} else if err := pipeWriter.Close(); err != nil {
		return fmt.Errorf("failed closing connecting pipe between Kustomize and YQ: %w", err)
	}

	// Wait for yq command to finish
	if err := yqCmd.Wait(); err != nil {
		return fmt.Errorf("failed running YQ command: %w", err)
	}

	return nil
}

func main() {

	// Create command structure
	cmd := command.MustNew(
		filepath.Base(os.Args[0]),
		"Devbot bake job prepares the resource manifest of a repository.",
		`This job prepares the Kubernetes resource manifest for a given repository
in preparation for deployment into an environment.'`,
		&Action{},
		[]any{
			&observability.LoggingHook{LogLevel: "info"},
			&observability.OTelHook{ServiceName: "devbot-bake-job"},
		},
	)

	os.Exit(int(command.Execute(os.Stderr, cmd, os.Args, command.EnvVarsArrayToMap(os.Environ()))))
}
