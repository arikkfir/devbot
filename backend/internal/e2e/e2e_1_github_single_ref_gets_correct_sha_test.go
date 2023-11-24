package e2e

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/controllers/repositories/github"
	"github.com/arikkfir/devbot/backend/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"slices"
	"testing"
	"time"
)

func TestGitHubRepoWithSingleRefGetsCorrectSHA(t *testing.T) {
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

	util.Eventually(t, 5*time.Minute, 2*time.Second, func(t util.TestingT) {
		r, err := k.GetGitHubRepository(ctx, repoCR.Namespace, repoCR.Name)
		if err != nil {
			t.Fatalf("Failed to get GitHubRepository object: %+v", err)
		}

		if !slices.Contains(r.Finalizers, github.RepositoryFinalizer) {
			t.Fatalf("GitHubRepository object does not contain finalizer '%s'", github.RepositoryFinalizer)
		}

		if r.DeletionTimestamp != nil {
			t.Fatalf("GitHubRepository object has deletion timestamp")
		}

		if c := r.GetStatusConditionCurrent(); c == nil || c.Status != metav1.ConditionTrue {
			t.Fatalf("GitHubRepository object %s condition is wrong: %+v", apiv1.ConditionTypeCurrent, c)
		}

		// TODO: search only for refs owned by our repo

		refs := &apiv1.GitHubRepositoryRefList{}
		if err := k.c.List(ctx, refs); err != nil {
			t.Fatalf("Failed getting list of GitHubRepositoryRef objects: %+v", err)
		}

		if len(refs.Items) != 1 {
			t.Fatalf("Incorrect number of GitHubRepositoryRef objects: %d", len(refs.Items))
		}

		ref := refs.Items[0]
		if ref.Spec.Ref != branchRefName {
			t.Fatalf("Incorrect GitHubRepositoryRef ref: expected '%s', was '%s'", branch, ref.Spec.Ref)
		}

		if ref.Status.CommitSHA != *commitResp.SHA {
			t.Fatalf("Incorrect GitHubRepositoryRef commit SHA: expected '%s', was '%s'", *commitResp.SHA, ref.Status.CommitSHA)
		}
	})
}
