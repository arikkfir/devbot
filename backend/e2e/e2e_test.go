package e2e_test

import (
	"context"
	"embed"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
)

var (
	//go:embed all:repositories/*
	repositoriesFS embed.FS
)

type E2E struct {
	Ctx context.Context
	GH  *testing.GClient
	K   *testing.KClient
}

func NewE2E(t T) *E2E {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return &E2E{
		Ctx: ctx,
		GH:  testing.GH(ctx, t),
		K:   testing.K(ctx, t),
	}
}

func (e *E2E) CreateGitHubAndK8sRepository(t T, ns *testing.KNamespace, name string, refreshInterval string) (*testing.GitHubRepositoryInfo, string) {
	ghRepo := e.GH.CreateRepository(t, repositoriesFS, "repositories/"+name)
	kRepoName := ns.CreateRepository(t, apiv1.RepositorySpec{
		GitHub: &apiv1.GitHubRepositorySpec{
			Owner:               ghRepo.Owner,
			Name:                ghRepo.Name,
			PersonalAccessToken: ns.CreateGitHubAuthSecretSpec(t, e.GH.Token, true),
		},
		RefreshInterval: refreshInterval,
	})
	return ghRepo, kRepoName
}
