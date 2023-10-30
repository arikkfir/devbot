package e2e

import (
	"context"
	"testing"
)

func TestSanity(t *testing.T) {
	ctx := context.Background()

	gh, err := NewGitHubTestClient(t, "devbot-testing")
	if err != nil {
		t.Fatalf("Failed to create GitHub client: %+v", err)
	}
	t.Cleanup(gh.Close)

	k8s, err := NewK8sTestClient(t, "default")
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %+v", err)
	}
	t.Cleanup(k8s.Close)

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
		t.Logf("Application: %+v", app)
	}

	if err := gh.CreateFile(ctx, *repo.Name, "main", "README.md", t.Name()); err != nil {
		t.Fatalf("Failed to create GitHub file: %+v", err)
	}

	// TODO: verify redis pubsub message sent

	if app, err := k8s.GetApplication(ctx, appName); err != nil {
		t.Fatalf("Failed to get application: %+v", err)
	} else {
		// TODO: verify application status was updated correctly
		t.Logf("Application: %+v", app)
	}
}
