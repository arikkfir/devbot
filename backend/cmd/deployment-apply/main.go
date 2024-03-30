package main

import (
	"context"
	"fmt"
	"github.com/arikkfir/devbot/backend/internal/util/configuration"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"path/filepath"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	devbotconfig "github.com/arikkfir/devbot/backend/internal/config"
)

const (
	kubectlBinaryFilePath = "/usr/local/bin/kubectl"
)

type Config struct {
	devbotconfig.CommandConfig
	ApplicationName string `env:"APPLICATION_OBJECT_NAME" long:"application" description:"Kubernetes Application object name" required:"true"`
	EnvironmentName string `env:"ENVIRONMENT_OBJECT_NAME" long:"environment" description:"Kubernetes Environment object name" required:"true"`
	DeploymentName  string `env:"DEPLOYMENT_OBJECT_NAME" long:"deployment" description:"Kubernetes Deployment object name" required:"true"`
	ManifestFile    string `env:"MANIFEST_FILE" long:"manifest-file" description:"File containing the generated Kubernetes resource manifests to apply" required:"true"`
}

var (
	cfg Config
)

func init() {
	configuration.Parse(&cfg)
	logging.Configure(os.Stderr, cfg.DevMode, cfg.LogLevel)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
		log.Fatal().Err(err).Msg("Failed starting kubectl command")
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal().Err(err).Msg("Failed running kubectl command")
	}
}
