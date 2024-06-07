package main

import (
	"context"
	"fmt"
	"github.com/arikkfir/command"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"os"
	"path/filepath"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type Action struct {
	Branch string `required:"true" desc:"Git branch to checkout."`
	GitURL string `required:"true" desc:"Git URL."`
	SHA    string `required:"true" desc:"Commit SHA to checkout."`
}

func (e *Action) Run(ctx context.Context) error {
	log.Logger = log.With().
		Str("gitURL", e.GitURL).
		Str("branch", e.Branch).
		Str("sha", e.SHA).
		Logger()

	// Calculate Git URL from repository
	if _, err := os.Stat("/data/.git"); errors.Is(err, os.ErrNotExist) {

		// Clone
		cloneOptions := &git.CloneOptions{
			URL:      e.GitURL,
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
	localBranchRefName := plumbing.NewBranchReferenceName(e.Branch)
	remoteBranchRefName := plumbing.NewRemoteReferenceName("origin", e.Branch)
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
	if err := worktree.Checkout(&git.CheckoutOptions{Force: true, Keep: false, Hash: plumbing.NewHash(e.SHA)}); err != nil {
		return errors.New("failed checking out revision: %w", err)
	}

	return nil
}

func main() {

	// Create command structure
	cmd := command.MustNew(
		filepath.Base(os.Args[0]),
		"Devbot clone job clones the given Git repository.",
		`This job clones the given Git repository.'`,
		&Action{},
		[]command.PreRunHook{&logging.InitHook{LogLevel: "info"}, &logging.SentryInitHook{}},
		[]command.PostRunHook{&logging.SentryFlushHook{}},
	)

	// Prepare a context that gets canceled if OS termination signals are sent
	ctx, cancel := context.WithCancel(command.SetupSignalHandler())
	defer cancel()

	// Execute the correct command
	os.Exit(int(command.Execute(ctx, os.Stderr, cmd, os.Args, command.EnvVarsArrayToMap(os.Environ()))))

}
