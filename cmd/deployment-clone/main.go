package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/arikkfir/command"
	"github.com/arikkfir/devbot/internal/util/observability"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"
	"os"
	"path"
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
	cloneOptions := &git.CloneOptions{
		URL:      e.GitURL,
		Progress: log.With().Str("process", "git").Logger(),
	}

	// Calculate Git URL from repository
	if _, err := os.Stat("/data/.git"); errors.Is(err, os.ErrNotExist) {
		if _, err := git.PlainClone("/data", false, cloneOptions); err != nil {
			return fmt.Errorf("failed cloning repository: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed inspecting target clone directory: %w", err)
	}

	// Open the cloned repository
	gitRepo, err := git.PlainOpen("/data")
	if err != nil {
		return fmt.Errorf("failed opening cloned repository: %w", err)
	}

	// Ensure
	remotes, err := gitRepo.Remotes()
	if err != nil {
		return fmt.Errorf("failed fetching remotes: %w", err)
	} else if len(remotes) != 1 {
		var remoteURLs []string
		for _, remote := range remotes {
			remoteURLs = append(remoteURLs, remote.Config().URLs...)
		}
		return fmt.Errorf("expected exactly 1 remotes, found %d: %+v", len(remotes), remoteURLs)
	} else if urls := remotes[0].Config().URLs; len(urls) != 1 {
		return fmt.Errorf("expected exactly 1 remote URLs, found %d: %+v", len(urls), urls)
	} else if url := urls[0]; url != e.GitURL {
		if entries, err := os.ReadDir("/data"); err != nil {
			return fmt.Errorf("failed reading directory: %w", err)
		} else {
			for _, entry := range entries {
				if err := os.RemoveAll(path.Join("/data", entry.Name())); err != nil {
					return fmt.Errorf("failed removing file: %w", err)
				}
			}
		}
		if _, err := git.PlainClone("/data", false, cloneOptions); err != nil {
			return fmt.Errorf("failed cloning repository: %w", err)
		}
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
		return fmt.Errorf("failed fetching branch: %w", err)
	}

	// Attempt to open the worktree
	worktree, err := gitRepo.Worktree()
	if err != nil {
		return fmt.Errorf("failed opening worktree: %w", err)
	}

	// Checkout the exact revision listed in the repository
	if err := worktree.Checkout(&git.CheckoutOptions{Force: true, Keep: false, Hash: plumbing.NewHash(e.SHA)}); err != nil {
		return fmt.Errorf("failed checking out revision: %w", err)
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
		[]any{
			&observability.LoggingHook{LogLevel: "info"},
			&observability.OTelHook{ServiceName: "devbot-clone-job"},
		},
	)

	// Execute the correct command
	os.Exit(int(command.Execute(os.Stderr, cmd, os.Args, command.EnvVarsArrayToMap(os.Environ()))))
}
