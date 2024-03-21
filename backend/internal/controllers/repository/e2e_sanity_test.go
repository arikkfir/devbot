package repository_test

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

type (
	conditionExpectations map[string]*metav1.Condition
	repositoryExpectation struct {
		conditions    conditionExpectations
		defaultBranch string
		revisions     map[string]string
	}
	repositoryExpectations map[string]repositoryExpectation
)

func validateRepositoryExpectations(t JustT, ctx context.Context, k *KClient, ns *KNamespace, expectations repositoryExpectations) {

	// Fetch repositories
	repositories := &apiv1.RepositoryList{}
	For(t).Expect(k.Client.List(ctx, repositories, client.InNamespace(ns.Name))).Will(Succeed())

	// Ensure each repository found corresponds to an expected repository
	for _, r := range repositories.Items {
		e, ok := expectations[r.Name]
		if !ok {
			t.Fatalf("Unexpected repository: %+v", r)
		}

		delete(expectations, r.Name)
		for conditionType, ce := range e.conditions {
			For(t).Expect(r.Status.GetCondition(conditionType)).Will(EqualCondition(ce))
		}
		For(t).Expect(r.Status.DefaultBranch).Will(BeEqualTo(e.defaultBranch))
		For(t).Expect(r.Status.Revisions).Will(BeEqualTo(e.revisions))
	}

	// Ensure each expected repository was indeed matched & found
	if len(expectations) > 0 {
		t.Fatalf("Missing repository expectations: %+v", expectations)
	}
}

func TestRepositoryReconciliation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	k := NewKubernetes(t)
	ns := k.CreateNamespace(t, ctx)
	gh := NewGitHub(t, ctx)

	ghCommonRepo := gh.CreateRepository(t, ctx, "app1/common")
	kCommonRepoName := ns.CreateRepository(t, ctx, apiv1.RepositorySpec{
		GitHub: &apiv1.GitHubRepositorySpec{
			Owner:               ghCommonRepo.Owner,
			Name:                ghCommonRepo.Name,
			PersonalAccessToken: ns.CreateGitHubAuthSecretSpec(t, ctx, gh.Token, true),
		},
		RefreshInterval: "10s",
	})

	ghServerRepo := gh.CreateRepository(t, ctx, "app1/server")
	kServerRepoName := ns.CreateRepository(t, ctx, apiv1.RepositorySpec{
		GitHub: &apiv1.GitHubRepositorySpec{
			Owner:               ghServerRepo.Owner,
			Name:                ghServerRepo.Name,
			PersonalAccessToken: ns.CreateGitHubAuthSecretSpec(t, ctx, gh.Token, true),
		},
		RefreshInterval: "10s",
	})

	// Validate initial reconciliation
	For(t).Expect(func(t JustT) {
		validateRepositoryExpectations(t, ctx, k, ns, repositoryExpectations{
			kCommonRepoName: {
				conditions: map[string]*metav1.Condition{
					apiv1.FailedToInitialize: nil,
					apiv1.Finalizing:         nil,
					apiv1.Invalid:            nil,
					apiv1.Stale:              nil,
					apiv1.Unauthenticated:    nil,
				},
				defaultBranch: "main",
				revisions:     map[string]string{"main": ghCommonRepo.GetBranchSHA(t, ctx, "main")},
			},
			kServerRepoName: {
				conditions: map[string]*metav1.Condition{
					apiv1.FailedToInitialize: nil,
					apiv1.Finalizing:         nil,
					apiv1.Invalid:            nil,
					apiv1.Stale:              nil,
					apiv1.Unauthenticated:    nil,
				},
				defaultBranch: "main",
				revisions:     map[string]string{"main": ghServerRepo.GetBranchSHA(t, ctx, "main")},
			},
		})
	}).Will(Eventually(Succeed()).Within(30 * time.Second).ProbingEvery(100 * time.Millisecond))

	// Create new branches
	commonRepoFeature1SHA := ghCommonRepo.CreateBranch(t, ctx, "feature1")
	serverRepoFeature2SHA := ghServerRepo.CreateBranch(t, ctx, "feature2")

	// Validate changes have been reconciled
	For(t).Expect(func(t JustT) {
		validateRepositoryExpectations(t, ctx, k, ns, repositoryExpectations{
			kCommonRepoName: {
				conditions: map[string]*metav1.Condition{
					apiv1.FailedToInitialize: nil,
					apiv1.Finalizing:         nil,
					apiv1.Invalid:            nil,
					apiv1.Stale:              nil,
					apiv1.Unauthenticated:    nil,
				},
				defaultBranch: "main",
				revisions: map[string]string{
					"main":     ghCommonRepo.GetBranchSHA(t, ctx, "main"),
					"feature1": commonRepoFeature1SHA,
				},
			},
			kServerRepoName: {
				conditions: map[string]*metav1.Condition{
					apiv1.FailedToInitialize: nil,
					apiv1.Finalizing:         nil,
					apiv1.Invalid:            nil,
					apiv1.Stale:              nil,
					apiv1.Unauthenticated:    nil,
				},
				defaultBranch: "main",
				revisions: map[string]string{
					"main":     ghServerRepo.GetBranchSHA(t, ctx, "main"),
					"feature2": serverRepoFeature2SHA,
				},
			},
		})
	}).Will(Eventually(Succeed()).Within(30 * time.Second).ProbingEvery(100 * time.Millisecond))

	// Create a new commit on the common repository
	commonRepoFeature1CommitSHA := ghCommonRepo.CreateFile(t, ctx, "feature1")

	// Validate changes have been reconciled
	For(t).Expect(func(t JustT) {
		validateRepositoryExpectations(t, ctx, k, ns, repositoryExpectations{
			kCommonRepoName: {
				conditions: map[string]*metav1.Condition{
					apiv1.FailedToInitialize: nil,
					apiv1.Finalizing:         nil,
					apiv1.Invalid:            nil,
					apiv1.Stale:              nil,
					apiv1.Unauthenticated:    nil,
				},
				defaultBranch: "main",
				revisions: map[string]string{
					"main":     ghCommonRepo.GetBranchSHA(t, ctx, "main"),
					"feature1": commonRepoFeature1CommitSHA,
				},
			},
			kServerRepoName: {
				conditions: map[string]*metav1.Condition{
					apiv1.FailedToInitialize: nil,
					apiv1.Finalizing:         nil,
					apiv1.Invalid:            nil,
					apiv1.Stale:              nil,
					apiv1.Unauthenticated:    nil,
				},
				defaultBranch: "main",
				revisions: map[string]string{
					"main":     ghServerRepo.GetBranchSHA(t, ctx, "main"),
					"feature2": serverRepoFeature2SHA,
				},
			},
		})
	}).Will(Eventually(Succeed()).Within(30 * time.Second).ProbingEvery(100 * time.Millisecond))
}
