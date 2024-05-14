package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/arikkfir/command"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-logr/logr"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/arikkfir/devbot/backend/internal/util/logging"

	"os"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Version represents the version of the controller. This variable gets its value by injection from the build process.
//
//goland:noinspection GoUnusedGlobalVariable
var Version = "0.0.0-unknown"

// Config is the configuration for this job.
type Config struct {
	DisableJSONLogging bool   `desc:"Disable JSON logging."`
	LogLevel           string `config:"required" desc:"Log level, must be one of: trace,debug,info,warn,error,fatal,panic"`
	Branch             string `config:"required" desc:"Git branch to checkout."`
	GitURL             string `config:"required" desc:"Git URL."`
	SHA                string `config:"required" desc:"Commit SHA to checkout."`
}

var rootCommand = command.New(command.Spec{
	Name:             filepath.Base(os.Args[0]),
	ShortDescription: "Devbot clone job clones the given Git repository.",
	LongDescription:  `This job clones the given Git repository.'`,
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
				return errors.New("failed cloning repository: %w", err)
			}

		} else if err != nil {
			return errors.New("failed inspecting target clone directory: %w", err)
		}

		// Open the cloned repository
		gitRepo, err := git.PlainOpen("/data")
		if err != nil {
			return errors.New("failed opening cloned repository: %w", err)
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
			return errors.New("failed fetching branch: %w", err)
		}

		// Attempt to open the worktree
		worktree, err := gitRepo.Worktree()
		if err != nil {
			return errors.New("failed opening worktree: %w", err)
		}

		// Checkout the exact revision listed in the repository
		if err := worktree.Checkout(&git.CheckoutOptions{Force: true, Keep: false, Hash: plumbing.NewHash(cfg.SHA)}); err != nil {
			return errors.New("failed checking out revision: %w", err)
		}

		return nil
	},
})

func main() {
	command.Execute(rootCommand, os.Args, command.EnvVarsArrayToMap(os.Environ()))
}
