package devctl

import (
	"context"
	"slices"

	"github.com/secureworks/errors"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/arikkfir/devbot/backend/internal/devctl/bootstrap"
)

var (
	// visibilities represents the allowed set of values for the GitHub repository "visibility" field
	visibilities = []string{"public", "private", "internal"}
)

// BootstrapGitHubExecutor is the command executor for the "devbot bootstrap github" command.
type BootstrapGitHubExecutor struct {
	Owner               string `required:"true" desc:"The owner of the repository to host the Devbot GitOps specification."`
	Name                string `required:"true" desc:"The name of the repository to host the Devbot GitOps specification."`
	Visibility          string `desc:"Repository visibility to use for the repository when created."`
	PersonalAccessToken string `required:"true" desc:"The personal access token to use for authentication."`
}

func (c *BootstrapGitHubExecutor) PreRun(_ context.Context) error {
	if !slices.Contains(visibilities, c.Visibility) {
		return errors.New("illegal visibility: %s", c.Visibility)
	}
	return nil
}

func (c *BootstrapGitHubExecutor) Run(ctx context.Context) error {

	restConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: clientcmd.RecommendedHomeFile},
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return errors.New("failed building Kubernetes client configuration: %w", err)
	}

	bootstrapper, err := bootstrap.NewGitHubBootstrapper(ctx, c.PersonalAccessToken, restConfig)
	if err != nil {
		return errors.New("failed to create GitHub Bootstrapper: " + err.Error())
	}

	if err := bootstrapper.Bootstrap(ctx, c.Owner, c.Name, c.Visibility); err != nil {
		return err
	}

	return nil
}
