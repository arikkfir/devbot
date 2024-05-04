package main

import (
	"context"
	"fmt"
	. "github.com/arikkfir/devbot/backend/internal/util/configuration"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"os"
	"os/exec"
	"path/filepath"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	disableJSONLoggingKey = "disable-json-logging"
	logLevelKey           = "log-level"
	applicationNameKey    = "application-name"
	environmentNameKey    = "environment-name"
	deploymentNameKey     = "deployment-name"
	manifestFileKey       = "manifest-file"
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
	DisableJSONLogging bool
	LogLevel           string
	ApplicationName    string
	EnvironmentName    string
	DeploymentName     string
	ManifestFile       string
}

// cfg is the configuration of the job. It is populated in the init function.
var cfg = Config{
	DisableJSONLogging: false,
	LogLevel:           "info",
}

func init() {

	// Configure & parse CLI flags
	pflag.BoolVar(&cfg.DisableJSONLogging, disableJSONLoggingKey, cfg.DisableJSONLogging, "Disable JSON logging")
	pflag.StringVar(&cfg.LogLevel, logLevelKey, cfg.LogLevel, "Log level, must be one of: trace,debug,info,warn,error,fatal,panic")
	pflag.StringVar(&cfg.ApplicationName, applicationNameKey, cfg.ApplicationName, "Kubernetes Application object name")
	pflag.StringVar(&cfg.EnvironmentName, environmentNameKey, cfg.EnvironmentName, "Kubernetes Environment object name")
	pflag.StringVar(&cfg.DeploymentName, deploymentNameKey, cfg.DeploymentName, "Kubernetes Deployment object name")
	pflag.StringVar(&cfg.ManifestFile, manifestFileKey, cfg.ManifestFile, "Target file to write resources YAML manifest to")
	pflag.Parse()

	// Allow the user to override configuration values using environment variables
	ApplyBoolEnvironmentVariableTo(&cfg.DisableJSONLogging, FlagNameToEnvironmentVariable(disableJSONLoggingKey))
	ApplyStringEnvironmentVariableTo(&cfg.LogLevel, FlagNameToEnvironmentVariable(logLevelKey))
	ApplyStringEnvironmentVariableTo(&cfg.ApplicationName, FlagNameToEnvironmentVariable(applicationNameKey))
	ApplyStringEnvironmentVariableTo(&cfg.EnvironmentName, FlagNameToEnvironmentVariable(environmentNameKey))
	ApplyStringEnvironmentVariableTo(&cfg.DeploymentName, FlagNameToEnvironmentVariable(deploymentNameKey))
	ApplyStringEnvironmentVariableTo(&cfg.ManifestFile, FlagNameToEnvironmentVariable(manifestFileKey))

	// Validate configuration
	if cfg.LogLevel == "" {
		log.Fatal().Msg("Log level cannot be empty")
	}
	if cfg.ApplicationName == "" {
		log.Fatal().Msg("Application name cannot be empty")
	}
	if cfg.EnvironmentName == "" {
		log.Fatal().Msg("Environment name cannot be empty")
	}
	if cfg.DeploymentName == "" {
		log.Fatal().Msg("Deployment name cannot be empty")
	}
	if cfg.ManifestFile == "" {
		log.Fatal().Msg("Manifest file name cannot be empty")
	}

	// Configure logging
	logging.Configure(os.Stderr, !cfg.DisableJSONLogging, cfg.LogLevel, Version)

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
