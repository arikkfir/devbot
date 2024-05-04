package main

import (
	"context"
	. "github.com/arikkfir/devbot/backend/internal/util/configuration"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"github.com/spf13/pflag"
	"io"
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
	actualBranchKey       = "actual-branch"
	applicationNameKey    = "application-name"
	baseDeployKey         = "base-deploy-dir"
	environmentNameKey    = "environment-name"
	deploymentNameKey     = "deployment-name"
	manifestFileKey       = "manifest-file"
	preferredBranchKey    = "preferred-branch"
	repoDefaultBranchKey  = "repo-default-branch"
	shaKey                = "sha"
)

const (
	// kustomizeBinaryFilePath is the path to the kustomize binary.
	kustomizeBinaryFilePath = "/usr/local/bin/kustomize"

	// yqBinaryFilePath is the path to the yq binary.
	yqBinaryFilePath = "/usr/local/bin/yq"
)

// Version represents the version of the controller. This variable gets its value by injection from the build process.
//
//goland:noinspection GoUnusedGlobalVariable
var Version = "0.0.0-unknown"

// Config is the configuration for this job.
type Config struct {
	DisableJSONLogging bool
	LogLevel           string
	ActualBranch       string
	ApplicationName    string
	BaseDeployDir      string
	EnvironmentName    string
	DeploymentName     string
	ManifestFile       string
	PreferredBranch    string
	RepoDefaultBranch  string
	SHA                string
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
	pflag.StringVar(&cfg.ActualBranch, actualBranchKey, cfg.ActualBranch, "Git branch serving as a default, in case the preferred branch was missing")
	pflag.StringVar(&cfg.ApplicationName, applicationNameKey, cfg.ApplicationName, "Kubernetes Application object name")
	pflag.StringVar(&cfg.BaseDeployDir, baseDeployKey, cfg.BaseDeployDir, "Base directory Directory holding the Kustomize overlay to build")
	pflag.StringVar(&cfg.EnvironmentName, environmentNameKey, cfg.EnvironmentName, "Kubernetes Environment object name")
	pflag.StringVar(&cfg.DeploymentName, deploymentNameKey, cfg.DeploymentName, "Kubernetes Deployment object name")
	pflag.StringVar(&cfg.ManifestFile, manifestFileKey, cfg.ManifestFile, "Target file to write resources YAML manifest to")
	pflag.StringVar(&cfg.PreferredBranch, preferredBranchKey, cfg.PreferredBranch, "Git branch preferred for baking, if it exists")
	pflag.StringVar(&cfg.RepoDefaultBranch, repoDefaultBranchKey, cfg.RepoDefaultBranch, "The default branch of the repository being deployed")
	pflag.StringVar(&cfg.SHA, shaKey, cfg.SHA, "Commit SHA to checkout")
	pflag.Parse()

	// Allow the user to override configuration values using environment variables
	ApplyBoolEnvironmentVariableTo(&cfg.DisableJSONLogging, FlagNameToEnvironmentVariable(disableJSONLoggingKey))
	ApplyStringEnvironmentVariableTo(&cfg.LogLevel, FlagNameToEnvironmentVariable(logLevelKey))
	ApplyStringEnvironmentVariableTo(&cfg.ActualBranch, FlagNameToEnvironmentVariable(actualBranchKey))
	ApplyStringEnvironmentVariableTo(&cfg.ApplicationName, FlagNameToEnvironmentVariable(applicationNameKey))
	ApplyStringEnvironmentVariableTo(&cfg.BaseDeployDir, FlagNameToEnvironmentVariable(baseDeployKey))
	ApplyStringEnvironmentVariableTo(&cfg.EnvironmentName, FlagNameToEnvironmentVariable(environmentNameKey))
	ApplyStringEnvironmentVariableTo(&cfg.DeploymentName, FlagNameToEnvironmentVariable(deploymentNameKey))
	ApplyStringEnvironmentVariableTo(&cfg.ManifestFile, FlagNameToEnvironmentVariable(manifestFileKey))
	ApplyStringEnvironmentVariableTo(&cfg.PreferredBranch, FlagNameToEnvironmentVariable(preferredBranchKey))
	ApplyStringEnvironmentVariableTo(&cfg.RepoDefaultBranch, FlagNameToEnvironmentVariable(repoDefaultBranchKey))
	ApplyStringEnvironmentVariableTo(&cfg.SHA, FlagNameToEnvironmentVariable(shaKey))

	// Validate configuration
	if cfg.LogLevel == "" {
		log.Fatal().Msg("Log level cannot be empty")
	}
	if cfg.ActualBranch == "" {
		log.Fatal().Msg("Actual branch cannot be empty")
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
	if cfg.PreferredBranch == "" {
		log.Fatal().Msg("Preferred branch name cannot be empty")
	}
	if cfg.RepoDefaultBranch == "" {
		log.Fatal().Msg("Repository default branch name cannot be empty")
	}

	// Configure logging
	logging.Configure(os.Stderr, !cfg.DisableJSONLogging, cfg.LogLevel, Version)

}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Logger = log.With().
		Str("actualBranch", cfg.ActualBranch).
		Str("appName", cfg.ApplicationName).
		Str("envName", cfg.EnvironmentName).
		Str("deploymentName", cfg.DeploymentName).
		Str("baseDeployDir", cfg.BaseDeployDir).
		Str("manifestFile", cfg.ManifestFile).
		Str("preferredBranch", cfg.PreferredBranch).
		Str("sha", cfg.SHA).
		Logger()

	// Create target resources file
	resourcesFile, err := os.Create(cfg.ManifestFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed creating target manifest file")
	}
	defer resourcesFile.Close()

	// Create a pipe that connects stdout of the "kustomize build" command to the "kustomize fn" command
	pipeReader, pipeWriter := io.Pipe()

	// This command produces resources from the kustomization file and outputs them to stdout
	var kustomizeCmd *exec.Cmd
	for _, path := range getKustomizeSearchPaths() {
		if _, err := os.Stat(path); err == nil {
			kustomizeCmd = exec.CommandContext(ctx, kustomizeBinaryFilePath, "build")
			kustomizeCmd.Dir = path
			break
		} else if !errors.Is(err, os.ErrNotExist) {
			log.Fatal().Err(err).Str("path", path).Msg("Failed inspecting path repository devbot path")
		}
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
		log.Fatal().Err(err).Msg("Failed starting command")
	}

	// This command accepts resources via stdin, processes them via the bash function script, and outputs to stdout
	yqCmd := exec.CommandContext(ctx, yqBinaryFilePath, `(.. | select(tag == "!!str")) |= envsubst`)
	yqCmd.Dir = kustomizeCmd.Dir
	yqCmd.Env = append(os.Environ(),
		"ACTUAL_BRANCH="+stringsutil.Slugify(cfg.ActualBranch),
		"APPLICATION="+stringsutil.Slugify(cfg.ApplicationName),
		"COMMIT_SHA="+cfg.SHA,
		"ENVIRONMENT="+stringsutil.Slugify(cfg.PreferredBranch),
		"PREFERRED_BRANCH="+stringsutil.Slugify(cfg.PreferredBranch),
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
		log.Fatal().Err(err).Msg("Failed starting command")
	}

	// Wait for kustomize command to finish
	if err := kustomizeCmd.Wait(); err != nil {
		log.Fatal().Err(err).Msg("Kustomize command failed")
	} else if err := pipeWriter.Close(); err != nil {
		log.Fatal().Err(err).Msg("Failed closing connecting pipe between Kustomize and YQ")
	}

	// Wait for yq command to finish
	if err := yqCmd.Wait(); err != nil {
		log.Fatal().Err(err).Msg("YQ command failed")
	}
}

func getKustomizeSearchPaths() []string {
	return []string{
		filepath.Join(cfg.BaseDeployDir, cfg.PreferredBranch),
		filepath.Join(cfg.BaseDeployDir, cfg.ActualBranch),
		filepath.Join(cfg.BaseDeployDir, cfg.RepoDefaultBranch),
		cfg.BaseDeployDir,
	}
}
