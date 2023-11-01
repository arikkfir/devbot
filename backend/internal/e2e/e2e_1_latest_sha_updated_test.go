package e2e

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAppRefLatestSHAIsUpdated(t *testing.T) {
	ctx := context.Background()
	gh := NewGitHubTestClient(t, "devbot-testing")
	k8s := NewK8sTestClient(t, "default")

	repo, err := gh.CreateRepository(ctx)
	if err != nil {
		t.Fatalf("Failed to create GitHub repository: %+v", err)
	}
	if err := gh.CreateRepositoryWebhook(ctx, *repo.Name, "push"); err != nil {
		t.Fatalf("Failed to create GitHub repository webhook: %+v", err)
	}

	var appName string
	if app, err := k8s.CreateApplication(ctx, gh.Owner, *repo.Name); err != nil {
		t.Fatalf("Failed to create application: %+v", err)
	} else {
		appName = app.Name
	}

	cr, err := gh.CreateFile(ctx, *repo.Name, "main", "README.md", t.Name())
	if err != nil {
		t.Fatalf("Failed to create GitHub file: %+v", err)
	}

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		app, err := k8s.GetApplication(ctx, appName)
		if err != nil {
			c.Errorf("Failed to get application: %+v", err)
			return
		}

		if app.Status.Refs == nil {
			c.Errorf("Application refs is nil")
			return
		}

		ref, ok := app.Status.Refs["refs/heads/main"]
		if !ok {
			c.Errorf("Application refs does not contain 'refs/heads/main'")
			return
		}

		if ref.LatestAvailableCommit != *cr.SHA {
			c.Errorf("Application latest SHA was not updated to '%s'", *cr.SHA)
			return
		}
	}, 30*time.Second, 2*time.Second)
}
