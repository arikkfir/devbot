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

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// kubectlBinaryFilePath is the path to the kubectl binary.
	kubectlBinaryFilePath = "/usr/local/bin/kubectl"
)

// Version represents the version of the controller. This variable gets its value by injection from the build process.
//
//goland:noinspection GoUnusedGlobalVariable
var Version = "0.0.0-unknown"

// Config is the configuration for this job.
type Config struct {
	DisableJSONLogging bool   `desc:"Disable JSON logging."`
	LogLevel           string `config:"required" desc:"Log level, must be one of: trace,debug,info,warn,error,fatal,panic"`
	ApplicationName    string `config:"required" desc:"Kubernetes Application object name."`
	EnvironmentName    string `config:"required" desc:"Kubernetes Environment object name."`
	DeploymentName     string `config:"required" desc:"Kubernetes Deployment object name."`
	ManifestFile       string `config:"required" desc:"Target file to write resources YAML manifest to."`
}

var rootCommand = command.New(command.Spec{
	Name:             filepath.Base(os.Args[0]),
	ShortDescription: "Devbot apply job deploys a pre-baked manifest to the cluster.",
	LongDescription: `This job applies, via 'kubectl', a pre-baked manifest to the
kubernetes cluster, thereby deploying a repository to a given environment.'`,
	Config: &Config{
		DisableJSONLogging: false,
		LogLevel:           "info",
	},
	Run: func(ctx context.Context, configAsAny any, usagePrinter command.UsagePrinter) error {
		cfg := configAsAny.(*Config)

		// Configure logging
		logging.Configure(os.Stderr, !cfg.DisableJSONLogging, cfg.LogLevel, Version)
		logrLogger := logr.New(&logging.ZeroLogLogrAdapter{}).V(0)
		ctrl.SetLogger(logrLogger)
		klog.SetLogger(logrLogger)
		log.Logger = log.With().
			Str("appName", cfg.ApplicationName).
			Str("envName", cfg.EnvironmentName).
			Str("deploymentName", cfg.DeploymentName).
			Str("outputManifest", cfg.ManifestFile).
			Logger()

		// Create the apply command
		cmd := exec.CommandContext(ctx, kubectlBinaryFilePath,
			"apply",
			fmt.Sprintf("--filename=%s", cfg.ManifestFile),
			fmt.Sprintf("--server-side=%v", true),
		)
		cmd.Dir = filepath.Dir(cfg.ManifestFile)
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
	},
})

func main() {
	command.Execute(rootCommand, os.Args, command.EnvVarsArrayToMap(os.Environ()))
}
