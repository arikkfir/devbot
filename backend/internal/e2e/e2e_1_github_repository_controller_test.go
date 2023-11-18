package e2e

import (
	"context"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/controllers"
	"github.com/arikkfir/devbot/backend/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"slices"
	"testing"
	"time"
)

func TestGitHubRepositoryCloning(t *testing.T) {
	ctx := context.Background()
	k := NewK8sTestClient(t)
	gh := NewGitHubTestClient(t, k.gitHubAuthSecrets["TOKEN"])

	ghRepo, err := gh.CreateRepository(ctx)
	if err != nil {
		t.Fatalf("Failed to create GitHub repository: %+v", err)
	}
	if err := gh.CreateRepositoryWebhook(ctx, *ghRepo.Name, k.gitHubAuthSecrets["WEBHOOK_SECRET"], "push"); err != nil {
		t.Fatalf("Failed to create GitHub repository webhook: %+v", err)
	}

	const branch = "main"
	var branchRefName = "refs/heads/" + branch
	commitResp, err := gh.CreateFile(ctx, *ghRepo.Name, branch, "README.md", t.Name())
	if err != nil {
		t.Fatalf("Failed to create GitHub file: %+v", err)
	}

	repoCR, err := k.CreateGitHubRepository(ctx, "default", *ghRepo.Owner.Login, *ghRepo.Name)
	if err != nil {
		t.Fatalf("Failed to create GitHubRepository object: %+v", err)
	}

	util.Eventually(t, 1*time.Minute, 2*time.Second, func(t util.TestingT) {
		r, err := k.GetGitHubRepository(ctx, repoCR.Namespace, repoCR.Name)
		if err != nil {
			t.Errorf("Failed to get GitHubRepository object: %+v", err)
			return
		}

		if !slices.Contains(r.Finalizers, controllers.GitHubRepositoryFinalizer) {
			t.Errorf("GitHubRepository object does not contain finalizer '%s'", controllers.GitHubRepositoryFinalizer)
			return
		}

		if r.DeletionTimestamp != nil {
			t.Fatalf("GitHubRepository object has deletion timestamp")
		}

		if r.Status.GitURL == "" {
			t.Errorf("GitHubRepository object has empty Git URL")
			return
		}

		expectedGitURL := fmt.Sprintf("https://github.com/%s/%s.git", *ghRepo.Owner.Login, *ghRepo.Name)
		if r.Status.GitURL != expectedGitURL {
			t.Errorf("GitHubRepository object has incorrect Git URL: expected '%s', was '%s'", expectedGitURL, r.Status.GitURL)
			return
		}

		if r.Status.LocalClonePath == "" {
			t.Errorf("GitHubRepository object has empty local clone path")
			return
		}

		if r.Status.LocalClonePath != fmt.Sprintf("/clones/%s/%s", *ghRepo.Owner.Login, *ghRepo.Name) {
			t.Errorf("GitHubRepository object has incorrect local clone path")
			return
		}

		if c := r.GetStatusConditionCloned(); c == nil || c.Status != metav1.ConditionTrue {
			t.Errorf("GitHubRepository object %s condition is wrong: %+v", apiv1.ConditionTypeCloned, c)
			return
		}
		if c := r.GetStatusConditionCurrent(); c == nil || c.Status != metav1.ConditionTrue {
			t.Errorf("GitHubRepository object %s condition is wrong: %+v", apiv1.ConditionTypeCurrent, c)
			return
		}

		refs := &apiv1.GitHubRepositoryRefList{}
		if err := k.c.List(ctx, refs); err != nil {
			t.Errorf("Failed getting list of GitHubRepositoryRef objects: %+v", err)
			return
		}

		if len(refs.Items) != 1 {
			t.Errorf("Incorrect number of GitHubRepositoryRef objects: %d", len(refs.Items))
			return
		}

		ref := refs.Items[0]
		if ref.Spec.Ref != branchRefName {
			t.Fatalf("Incorrect GitHubRepositoryRef ref: expected '%s', was '%s'", branch, ref.Spec.Ref)
			return
		}

		if ref.Status.CommitSHA != *commitResp.SHA {
			t.Errorf("Incorrect GitHubRepositoryRef commit SHA: expected '%s', was '%s'", *commitResp.SHA, ref.Status.CommitSHA)
			return
		}
	})
}
