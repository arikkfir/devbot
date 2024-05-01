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
	e2e := NewE2E(t)
	ns := e2e.K.CreateNamespace(e2e.Ctx, t)

	ghCommonRepo := e2e.GH.CreateRepository(e2e.Ctx, t, repositoriesFS, "repositories/common")
	kCommonRepoName := ns.CreateRepository(e2e.Ctx, t, apiv1.RepositorySpec{
		GitHub: &apiv1.GitHubRepositorySpec{
			Owner:               ghCommonRepo.Owner,
			Name:                ghCommonRepo.Name,
			PersonalAccessToken: ns.CreateGitHubAuthSecretSpec(e2e.Ctx, t, e2e.GH.Token, true),
		},
		RefreshInterval: "5s",
	})

	ghServerRepo := e2e.GH.CreateRepository(e2e.Ctx, t, repositoriesFS, "repositories/server")
	kServerRepoName := ns.CreateRepository(e2e.Ctx, t, apiv1.RepositorySpec{
		GitHub: &apiv1.GitHubRepositorySpec{
			Owner:               ghServerRepo.Owner,
			Name:                ghServerRepo.Name,
			PersonalAccessToken: ns.CreateGitHubAuthSecretSpec(e2e.Ctx, t, e2e.GH.Token, true),
		},
		RefreshInterval: "5s",
	})

	// Validate initial reconciliation
	With(t).Verify(func(t T) {
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
					Revisions:     map[string]string{"main": ghCommonRepo.GetBranchSHA(e2e.Ctx, t, "main")},
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
					Revisions:     map[string]string{"main": ghServerRepo.GetBranchSHA(e2e.Ctx, t, "main")},
				},
			},
		}
		reposList := &apiv1.RepositoryList{}
		With(t).Verify(K(e2e.Ctx, t).Client.List(e2e.Ctx, reposList, client.InNamespace(ns.Name))).Will(Succeed()).OrFail()
		With(t).Verify(reposList.Items).Will(EqualTo(repositoryExpectations).Using(RepositoriesComparator)).OrFail()
	}).Will(Succeed()).Within(10*time.Second, 100*time.Millisecond)

	// Create new branches
	commonRepoFeature1SHA := ghCommonRepo.CreateBranch(e2e.Ctx, t, "feature1")
	serverRepoFeature2SHA := ghServerRepo.CreateBranch(e2e.Ctx, t, "feature2")

	// Validate changes have been reconciled
	With(t).Verify(func(t T) {
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
					Revisions: map[string]string{
						"main":     ghCommonRepo.GetBranchSHA(e2e.Ctx, t, "main"),
						"feature1": commonRepoFeature1SHA,
					},
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
					Revisions: map[string]string{
						"main":     ghServerRepo.GetBranchSHA(e2e.Ctx, t, "main"),
						"feature2": serverRepoFeature2SHA,
					},
				},
			},
		}
		reposList := &apiv1.RepositoryList{}
		With(t).Verify(e2e.K.Client.List(e2e.Ctx, reposList, client.InNamespace(ns.Name))).Will(Succeed()).OrFail()
		With(t).Verify(reposList.Items).Will(EqualTo(repositoryExpectations).Using(RepositoriesComparator)).OrFail()

	}).Will(Succeed()).Within(20*time.Second, 100*time.Millisecond)

	// Create a new commit on the common repository
	commonRepoFeature1CommitSHA := ghCommonRepo.CreateFile(e2e.Ctx, t, "feature1")

	// Validate changes have been reconciled
	With(t).Verify(func(t T) {
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
					Revisions: map[string]string{
						"main":     ghCommonRepo.GetBranchSHA(e2e.Ctx, t, "main"),
						"feature1": commonRepoFeature1CommitSHA,
					},
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
					Revisions: map[string]string{
						"main":     ghServerRepo.GetBranchSHA(e2e.Ctx, t, "main"),
						"feature2": ghServerRepo.GetBranchSHA(e2e.Ctx, t, "feature2"),
					},
				},
			},
		}
		reposList := &apiv1.RepositoryList{}
		With(t).Verify(e2e.K.Client.List(e2e.Ctx, reposList, client.InNamespace(ns.Name))).Will(Succeed()).OrFail()
		With(t).Verify(reposList.Items).Will(EqualTo(repositoryExpectations).Using(RepositoriesComparator)).OrFail()

	}).Will(Succeed()).Within(10*time.Second, 100*time.Millisecond)
}
