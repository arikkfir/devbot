package devctl

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/secureworks/errors"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/arikkfir/devbot/backend/internal/devctl/bootstrap"
)

var (
	// visibilities represents the allowed set of values for the GitHub repository "visibility" field
	visibilities = []string{"public", "private", "internal"}
)

// GitHubBootstrapAction is the command executor for the "devbot bootstrap github" command.
type GitHubBootstrapAction struct {
	Owner               string `required:"true" desc:"The owner of the repository to host the Devbot GitOps specification."`
	Name                string `required:"true" desc:"The name of the repository to host the Devbot GitOps specification."`
	Visibility          string `desc:"Repository visibility to use for the repository when created."`
	PersonalAccessToken string `required:"true" desc:"The personal access token to use for authentication."`
	Timeout             string `desc:"How long to wait for Devbot to become ready (defaults to 10m.)"`
	GithubWebhooksURL   string `desc:"Webhooks URL to set in GitHub repositories with webhook configuration. If not specified, webhooks functionality is disabled, and repositories with webhook configuration will transition to the Invalid condition."`
}

func (c *GitHubBootstrapAction) Run(ctx context.Context) error {
	if !slices.Contains(visibilities, c.Visibility) {
		return errors.New("illegal visibility: %s", c.Visibility)
	}

	restConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: clientcmd.RecommendedHomeFile},
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return errors.New("failed building Kubernetes client configuration: %w", err)
	}

	if c.Timeout == "" {
		c.Timeout = "10m"
	}
	timeout, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout duration: %w", err)
	}

	bootstrapper, err := bootstrap.NewGitHubBootstrapper(ctx, c.PersonalAccessToken, timeout, c.GithubWebhooksURL, restConfig)
	if err != nil {
		return errors.New("failed to create GitHub Bootstrapper: " + err.Error())
	}

	if err := bootstrapper.Bootstrap(ctx, c.Owner, c.Name, c.Visibility); err != nil {
		return err
	}

	return nil
}
