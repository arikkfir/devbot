package e2e_test

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/e2e/expectations"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

func TestRepositoryReconciliation(t *testing.T) {
	t.Parallel()

	ns := K(t).CreateNamespace(t)

	ghCommonRepo := GH(t).CreateRepository(t, repositoriesFS, "repositories/common")
	kCommonRepoName := ns.CreateRepository(t, apiv1.RepositorySpec{
		GitHub: &apiv1.GitHubRepositorySpec{
			Owner:               ghCommonRepo.Owner,
			Name:                ghCommonRepo.Name,
			PersonalAccessToken: ns.CreateGitHubAuthSecretSpec(t, GH(t).Token, true),
		},
		RefreshInterval: "10s",
	})

	ghServerRepo := GH(t).CreateRepository(t, repositoriesFS, "repositories/server")
	kServerRepoName := ns.CreateRepository(t, apiv1.RepositorySpec{
		GitHub: &apiv1.GitHubRepositorySpec{
			Owner:               ghServerRepo.Owner,
			Name:                ghServerRepo.Name,
			PersonalAccessToken: ns.CreateGitHubAuthSecretSpec(t, GH(t).Token, true),
		},
		RefreshInterval: "10s",
	})

	// Validate initial reconciliation
	For(t).Expect(func(t TT) {
		repositoryExpectations := []RepositoryE{
			{
				Name: kCommonRepoName,
				Status: RepositoryStatusE{
					Conditions: map[string]*ConditionE{
						apiv1.FailedToInitialize: nil,
						apiv1.Finalizing:         nil,
						apiv1.Invalid:            nil,
						apiv1.Stale:              nil,
						apiv1.Unauthenticated:    nil,
					},
					DefaultBranch: "main",
					Revisions:     map[string]string{"main": ghCommonRepo.GetBranchSHA(t, "main")},
				},
			},
			{
				Name: kServerRepoName,
				Status: RepositoryStatusE{
					Conditions: map[string]*ConditionE{
						apiv1.FailedToInitialize: nil,
						apiv1.Finalizing:         nil,
						apiv1.Invalid:            nil,
						apiv1.Stale:              nil,
						apiv1.Unauthenticated:    nil,
					},
					DefaultBranch: "main",
					Revisions:     map[string]string{"main": ghServerRepo.GetBranchSHA(t, "main")},
				},
			},
		}
		reposList := &apiv1.RepositoryList{}
		For(t).Expect(K(t).Client.List(t, reposList, client.InNamespace(ns.Name))).Will(Succeed())
		For(t).Expect(reposList.Items).Will(CompareTo(repositoryExpectations).Using(RepositoriesComparator))
	}).Will(Eventually(Succeed()).Within(30 * time.Second).ProbingEvery(100 * time.Millisecond))

	//// Create new branches
	//commonRepoFeature1SHA := ghCommonRepo.CreateBranch(t, ctx, "feature1")
	//serverRepoFeature2SHA := ghServerRepo.CreateBranch(t, ctx, "feature2")
	//
	//// Validate changes have been reconciled
	//For(t).Expect(func(t JustT) {
	//	validateRepositoryExpectations(t, ctx, k, ns, repositoryExpectations{
	//		kCommonRepoName: {
	//			conditions: map[string]*metav1.Condition{
	//				apiv1.FailedToInitialize: nil,
	//				apiv1.Finalizing:         nil,
	//				apiv1.Invalid:            nil,
	//				apiv1.Stale:              nil,
	//				apiv1.Unauthenticated:    nil,
	//			},
	//			defaultBranch: "main",
	//			revisions: map[string]string{
	//				"main":     ghCommonRepo.GetBranchSHA(t, ctx, "main"),
	//				"feature1": commonRepoFeature1SHA,
	//			},
	//		},
	//		kServerRepoName: {
	//			conditions: map[string]*metav1.Condition{
	//				apiv1.FailedToInitialize: nil,
	//				apiv1.Finalizing:         nil,
	//				apiv1.Invalid:            nil,
	//				apiv1.Stale:              nil,
	//				apiv1.Unauthenticated:    nil,
	//			},
	//			defaultBranch: "main",
	//			revisions: map[string]string{
	//				"main":     ghServerRepo.GetBranchSHA(t, ctx, "main"),
	//				"feature2": serverRepoFeature2SHA,
	//			},
	//		},
	//	})
	//}).Will(Eventually(Succeed()).Within(30 * time.Second).ProbingEvery(100 * time.Millisecond))
	//
	//// Create a new commit on the common repository
	//commonRepoFeature1CommitSHA := ghCommonRepo.CreateFile(t, ctx, "feature1")
	//
	//// Validate changes have been reconciled
	//For(t).Expect(func(t JustT) {
	//	validateRepositoryExpectations(t, ctx, k, ns, repositoryExpectations{
	//		kCommonRepoName: {
	//			conditions: map[string]*metav1.Condition{
	//				apiv1.FailedToInitialize: nil,
	//				apiv1.Finalizing:         nil,
	//				apiv1.Invalid:            nil,
	//				apiv1.Stale:              nil,
	//				apiv1.Unauthenticated:    nil,
	//			},
	//			defaultBranch: "main",
	//			revisions: map[string]string{
	//				"main":     ghCommonRepo.GetBranchSHA(t, ctx, "main"),
	//				"feature1": commonRepoFeature1CommitSHA,
	//			},
	//		},
	//		kServerRepoName: {
	//			conditions: map[string]*metav1.Condition{
	//				apiv1.FailedToInitialize: nil,
	//				apiv1.Finalizing:         nil,
	//				apiv1.Invalid:            nil,
	//				apiv1.Stale:              nil,
	//				apiv1.Unauthenticated:    nil,
	//			},
	//			defaultBranch: "main",
	//			revisions: map[string]string{
	//				"main":     ghServerRepo.GetBranchSHA(t, ctx, "main"),
	//				"feature2": serverRepoFeature2SHA,
	//			},
	//		},
	//	})
	//}).Will(Eventually(Succeed()).Within(30 * time.Second).ProbingEvery(100 * time.Millisecond))
}
