package e2e_test

import (
	"embed"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
)

var (
	//go:embed all:repositories/*
	repositoriesFS embed.FS
)

func createGitHubAndK8sRepository(t T, ns *KNamespace, name string) (*GitHubRepositoryInfo, string) {
	GetHelper(t).Helper()
	ghRepo := GH(t).CreateRepository(t, repositoriesFS, "repositories/"+name)
	kRepoName := ns.CreateRepository(t, apiv1.RepositorySpec{
		GitHub: &apiv1.GitHubRepositorySpec{
			Owner:               ghRepo.Owner,
			Name:                ghRepo.Name,
			PersonalAccessToken: ns.CreateGitHubAuthSecretSpec(t, GH(t).Token, true),
		},
		RefreshInterval: "30s",
	})
	return ghRepo, kRepoName
}
