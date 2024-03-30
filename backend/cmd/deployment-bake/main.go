package main

import (
	"context"
	"github.com/arikkfir/devbot/backend/internal/util/configuration"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	devbotconfig "github.com/arikkfir/devbot/backend/internal/config"
)

const (
	kustomizeBinaryFilePath = "/usr/local/bin/kustomize"
	yqBinaryFilePath        = "/usr/local/bin/yq"
)

type Config struct {
	devbotconfig.CommandConfig
	ActualBranch      string `env:"ACTUAL_BRANCH" long:"actual-branch" description:"Git branch serving as a default, in case the preferred branch was missing" required:"true"`
	ApplicationName   string `env:"APPLICATION_OBJECT_NAME" long:"application" description:"Kubernetes Application object name" required:"true"`
	BaseDeployDir     string `env:"BASE_DEPLOY_DIR" long:"base-deploy-dir" description:"Base directory Directory holding the Kustomize overlay to build" required:"true"`
	EnvironmentName   string `env:"ENVIRONMENT_OBJECT_NAME" long:"environment" description:"Kubernetes Environment object name" required:"true"`
	DeploymentName    string `env:"DEPLOYMENT_OBJECT_NAME" long:"deployment" description:"Kubernetes Deployment object name" required:"true"`
	ManifestFile      string `env:"MANIFEST_FILE" long:"manifest-file" description:"Target file to write resources YAML manifest to" required:"true"`
	PreferredBranch   string `env:"PREFERRED_BRANCH" long:"preferred-branch" description:"Git branch preferred for baking, if it exists" required:"true"`
	RepoDefaultBranch string `env:"REPO_DEFAULT_BRANCH" long:"repo-default-branch" description:"The default branch of the repository being deployed" required:"true"`
	SHA               string `env:"SHA" long:"sha" description:"Commit SHA to checkout" required:"true"`
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
	resourcesFile, err := os.Create(filepath.Join("/data", cfg.ManifestFile))
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
