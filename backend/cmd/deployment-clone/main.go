package main

import (
	"context"
	"fmt"
	. "github.com/arikkfir/devbot/backend/internal/util/configuration"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"github.com/spf13/pflag"
	"os"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	disableJSONLoggingKey = "disable-json-logging"
	logLevelKey           = "log-level"
	branchKey             = "branch"
	gitURLKey             = "git-url"
	shaKey                = "sha"
)

// Version represents the version of the controller. This variable gets its value by injection from the build process.
//
//goland:noinspection GoUnusedGlobalVariable
var Version = "0.0.0-unknown"

// Config is the configuration for this job.
type Config struct {
	DisableJSONLogging bool
	LogLevel           string
	Branch             string
	GitURL             string
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
	pflag.StringVar(&cfg.Branch, branchKey, cfg.Branch, "Git branch to checkout")
	pflag.StringVar(&cfg.GitURL, gitURLKey, cfg.GitURL, "Git URL")
	pflag.StringVar(&cfg.SHA, shaKey, cfg.SHA, "Commit SHA to checkout")
	pflag.Parse()

	// Allow the user to override configuration values using environment variables
	ApplyBoolEnvironmentVariableTo(&cfg.DisableJSONLogging, FlagNameToEnvironmentVariable(disableJSONLoggingKey))
	ApplyStringEnvironmentVariableTo(&cfg.LogLevel, FlagNameToEnvironmentVariable(logLevelKey))
	ApplyStringEnvironmentVariableTo(&cfg.Branch, FlagNameToEnvironmentVariable(branchKey))
	ApplyStringEnvironmentVariableTo(&cfg.GitURL, FlagNameToEnvironmentVariable(gitURLKey))
	ApplyStringEnvironmentVariableTo(&cfg.SHA, FlagNameToEnvironmentVariable(shaKey))

	// Validate configuration
	if cfg.LogLevel == "" {
		log.Fatal().Msg("Log level cannot be empty")
	}
	if cfg.Branch == "" {
		log.Fatal().Msg("Branch cannot be empty")
	}
	if cfg.GitURL == "" {
		log.Fatal().Msg("Git URL cannot be empty")
	}
	if cfg.SHA == "" {
		log.Fatal().Msg("SHA cannot be empty")
	}

	// Configure logging
	logging.Configure(os.Stderr, !cfg.DisableJSONLogging, cfg.LogLevel, Version)

}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Logger = log.With().
		Str("gitURL", cfg.GitURL).
		Str("branch", cfg.Branch).
		Str("sha", cfg.SHA).
		Logger()

	// Calculate Git URL from repository
	if _, err := os.Stat("/data/.git"); errors.Is(err, os.ErrNotExist) {

		// Clone
		cloneOptions := &git.CloneOptions{
			URL:      cfg.GitURL,
			Progress: log.With().Str("process", "git").Logger(),
		}
		if _, err := git.PlainClone("/data", false, cloneOptions); err != nil {
			log.Fatal().Err(err).Msg("Failed cloning repository")
		}

	} else if err != nil {
		log.Fatal().Err(err).Msg("Failed inspecting target clone directory")
	}

	// Open the cloned repository
	gitRepo, err := git.PlainOpen("/data")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed opening cloned repository")
	}

	// Fetch our branch
	localBranchRefName := plumbing.NewBranchReferenceName(cfg.Branch)
	remoteBranchRefName := plumbing.NewRemoteReferenceName("origin", cfg.Branch)
	refSpec := fmt.Sprintf("%s:%s", localBranchRefName, remoteBranchRefName)
	fetchOptions := git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec(refSpec)},
		Progress:   log.With().Str("process", "git").Logger(),
	}
	if err := gitRepo.FetchContext(ctx, &fetchOptions); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		log.Fatal().Err(err).Msg("Failed fetching branch")
	}

	// Attempt to open the worktree
	worktree, err := gitRepo.Worktree()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed opening worktree")
	}

	// Checkout the exact revision listed in the repository
	if err := worktree.Checkout(&git.CheckoutOptions{Force: true, Keep: false, Hash: plumbing.NewHash(cfg.SHA)}); err != nil {
		log.Fatal().Err(err).Msg("Failed checking out revision")
	}
}
