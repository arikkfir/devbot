package main

import (
	"context"

	"github.com/arikkfir/command"
	"github.com/go-logr/logr"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/arikkfir/devbot/backend/internal/util/logging"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"

	"io"
	"os"
	"os/exec"
	"path/filepath"

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

// Version represents the version of the controller. This variable gets its value by injection from the build process.
//
//goland:noinspection GoUnusedGlobalVariable
var Version = "0.0.0-unknown"

// Config is the configuration for this job.
type Config struct {
	DisableJSONLogging bool   `desc:"Disable JSON logging."`
	LogLevel           string `config:"required" desc:"Log level, must be one of: trace,debug,info,warn,error,fatal,panic"`
	ActualBranch       string `config:"required" desc:"Git branch serving as a default, in case the preferred branch was missing."`
	ApplicationName    string `config:"required" desc:"Kubernetes Application object name."`
	BaseDeployDir      string `config:"required" desc:"Base directory Directory holding the Kustomize overlay to build."`
	EnvironmentName    string `config:"required" desc:"Kubernetes Environment object name."`
	DeploymentName     string `config:"required" desc:"Kubernetes Deployment object name."`
	ManifestFile       string `config:"required" desc:"Target file to write resources YAML manifest to."`
	PreferredBranch    string `config:"required" desc:"Git branch preferred for baking, if it exists."`
	RepoDefaultBranch  string `config:"required" desc:"The default branch of the repository being deployed."`
	SHA                string `config:"required" desc:"Commit SHA to checkout."`
}

var rootCommand = command.New(command.Spec{
	Name:             filepath.Base(os.Args[0]),
	ShortDescription: "Devbot bake job prepares the resource manifest of a repository.",
	LongDescription: `This job prepares the Kubernetes resource manifest for a given repository
in preparation for deployment into an environment.'`,
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
			return errors.New("failed creating target manifest file: %w", err)
		}
		defer resourcesFile.Close()

		// Create a pipe that connects stdout of the "kustomize build" command to the "kustomize fn" command
		pipeReader, pipeWriter := io.Pipe()

		// This command produces resources from the kustomization file and outputs them to stdout
		var kustomizeCmd *exec.Cmd
		searchPaths := []string{
			filepath.Join(cfg.BaseDeployDir, cfg.PreferredBranch),
			filepath.Join(cfg.BaseDeployDir, cfg.ActualBranch),
			filepath.Join(cfg.BaseDeployDir, cfg.RepoDefaultBranch),
			cfg.BaseDeployDir,
		}
		for _, path := range searchPaths {
			if _, err := os.Stat(path); err == nil {
				kustomizeCmd = exec.CommandContext(ctx, kustomizeBinaryFilePath, "build")
				kustomizeCmd.Dir = path
				break
			} else if !errors.Is(err, os.ErrNotExist) {
				return errors.New("failed inspecting path repository devbot path: %w", err)
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
			return errors.New("failed starting kustomize command: %w", err)
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
			return errors.New("failed starting yq command: %w", err)
		}

		// Wait for kustomize command to finish
		if err := kustomizeCmd.Wait(); err != nil {
			return errors.New("failed running kustomize command: %w", err)
		} else if err := pipeWriter.Close(); err != nil {
			return errors.New("failed closing connecting pipe between Kustomize and YQ: %w", err)
		}

		// Wait for yq command to finish
		if err := yqCmd.Wait(); err != nil {
			return errors.New("failed running YQ command: %w", err)
		}

		return nil
	},
})

func main() {
	command.Execute(rootCommand, os.Args, command.EnvVarsArrayToMap(os.Environ()))
}
