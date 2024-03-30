package main

import (
	"context"
	"fmt"
	"github.com/arikkfir/devbot/backend/internal/util/configuration"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"os"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	devbotconfig "github.com/arikkfir/devbot/backend/internal/config"
)

type Config struct {
	devbotconfig.CommandConfig
	Branch string `env:"BRANCH" long:"branch" description:"Git branch to checkout" required:"true"`
	GitURL string `env:"GIT_URL" long:"git-url" description:"Git URL" required:"true"`
	SHA    string `env:"SHA" long:"sha" description:"Commit SHA to checkout" required:"true"`
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
