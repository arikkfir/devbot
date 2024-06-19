package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/arikkfir/command"
	"github.com/rs/zerolog/log"

	"github.com/arikkfir/devbot/internal/util/observability"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// kubectlBinaryFilePath is the path to the kubectl binary.
	kubectlBinaryFilePath = "/usr/local/bin/kubectl"
)

type Action struct {
	ApplicationName string `required:"true" desc:"Kubernetes Application object name."`
	EnvironmentName string `required:"true" desc:"Kubernetes Environment object name."`
	DeploymentName  string `required:"true" desc:"Kubernetes Deployment object name."`
	ManifestFile    string `required:"true" desc:"Target file to write resources YAML manifest to."`
}

func (e *Action) Run(ctx context.Context) error {
	log.Logger = log.With().
		Str("appName", e.ApplicationName).
		Str("envName", e.EnvironmentName).
		Str("deploymentName", e.DeploymentName).
		Str("outputManifest", e.ManifestFile).
		Logger()

	// Create the apply command
	cmd := exec.CommandContext(ctx, kubectlBinaryFilePath,
		"apply",
		fmt.Sprintf("--filename=%s", e.ManifestFile),
		fmt.Sprintf("--server-side=%v", true),
	)
	cmd.Dir = filepath.Dir(e.ManifestFile)
	cmd.Stderr = log.With().Str("output", "stderr").Logger()
	cmd.Stdout = log.With().Str("output", "stdout").Logger()

	log.Info().Str("command", cmd.String()).Msg("Running kubectl command")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed starting kubectl command: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed running kubectl command: %w", err)
	}

	return nil
}

func main() {

	// Create command structure
	cmd := command.MustNew(
		filepath.Base(os.Args[0]),
		"Devbot apply job deploys a pre-baked manifest to the cluster.",
		`This job applies, via 'kubectl', a pre-baked manifest to the
kubernetes cluster, thereby deploying a repository to a given environment.'`,
		&Action{},
		[]any{
			&observability.LoggingHook{LogLevel: "info"},
			&observability.OTelHook{ServiceName: "devbot-apply-job"},
		},
	)

	os.Exit(int(command.Execute(os.Stderr, cmd, os.Args, command.EnvVarsArrayToMap(os.Environ()))))
}
