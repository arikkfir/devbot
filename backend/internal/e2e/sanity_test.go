package e2e

import (
	"context"
	"testing"
	"time"
)

func TestSanity(t *testing.T) {
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

	ctx := context.Background()

	repo, err := gh.CreateRepository(ctx)
	if err != nil {
		t.Fatalf("Failed to create GitHub repository: %+v", err)
	}
	if err := gh.CreateRepositoryWebhook(ctx, *repo.Name, "push"); err != nil {
		t.Fatalf("Failed to create GitHub repository webhook: %+v", err)
	}

	t.Log("Testing testing 1 2 3")
	time.Sleep(10 * time.Second)
}
