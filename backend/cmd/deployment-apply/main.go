package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/arikkfir/command"
	"github.com/go-logr/logr"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/arikkfir/devbot/backend/internal/util/version"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// kubectlBinaryFilePath is the path to the kubectl binary.
	kubectlBinaryFilePath = "/usr/local/bin/kubectl"
)

type Executor struct {
	DisableJSONLogging bool   `desc:"Disable JSON logging."`
	LogLevel           string `required:"true" desc:"Log level, must be one of: trace,debug,info,warn,error,fatal,panic"`
	ApplicationName    string `required:"true" desc:"Kubernetes Application object name."`
	EnvironmentName    string `required:"true" desc:"Kubernetes Environment object name."`
	DeploymentName     string `required:"true" desc:"Kubernetes Deployment object name."`
	ManifestFile       string `required:"true" desc:"Target file to write resources YAML manifest to."`
}

func (e *Executor) PreRun(_ context.Context) error { return nil }
func (e *Executor) Run(ctx context.Context) error {

	// Configure logging
	logging.Configure(os.Stderr, !e.DisableJSONLogging, e.LogLevel, version.Version)
	logrLogger := logr.New(&logging.ZeroLogLogrAdapter{}).V(0)
	ctrl.SetLogger(logrLogger)
	klog.SetLogger(logrLogger)
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
		return errors.New("failed starting kubectl command: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return errors.New("failed running kubectl command: %w", err)
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
		&Executor{
			DisableJSONLogging: false,
			LogLevel:           "info",
		},
	)

	// Prepare a context that gets canceled if OS termination signals are sent
	ctx, cancel := context.WithCancel(command.SetupSignalHandler())
	defer cancel()

	// Execute the correct command
	command.Execute(ctx, os.Stderr, cmd, os.Args, command.EnvVarsArrayToMap(os.Environ()))

}
