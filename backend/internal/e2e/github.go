package e2e

import (
	"context"
	"github.com/arikkfir/devbot/backend/internal/util"
	"github.com/google/go-github/v56/github"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
)

type GitHubTestClient struct {
	Owner   string
	client  *github.Client
	cleanup []func() error
}

func NewGitHubTestClient(owner string, token string) (*GitHubTestClient, error) {
	return &GitHubTestClient{
		Owner:  owner,
		client: github.NewClient(nil).WithAuthToken(token),
	}, nil
}

func (c *GitHubTestClient) CreateRepository(ctx context.Context) (*github.Repository, error) {
	repo, _, err := c.client.Repositories.Create(ctx, c.Owner, &github.Repository{
		Name:        &[]string{"devbot-test-" + util.RandomHash(7)}[0],
		Description: &[]string{"This is a test repository created by the devbot setup script"}[0],
		Topics:      []string{"devbot", "test"},
		Visibility:  &[]string{"public"}[0],
	})
	if err != nil {
		return nil, errors.New("failed to create test repository", err)
	}
	c.cleanup = append(c.cleanup, func() error {
		log.Info().Str("repo", *repo.FullName).Msg("Deleting repository")
		_, err := c.client.Repositories.Delete(ctx, c.Owner, *repo.Name)
		if err != nil {
			return errors.New("failed to delete test repository", errors.Meta("repo", *repo.FullName), err)
		}
		return nil
	})
	return repo, nil
}

func (c *GitHubTestClient) CreateRepositoryWebhook(ctx context.Context, repoName string, events ...string) error {
	_, _, err := c.client.Repositories.CreateHook(ctx, c.Owner, repoName, &github.Hook{
		Name:   &[]string{"web"}[0],
		Active: &[]bool{true}[0],
		Events: events,
		Config: map[string]interface{}{
			"url":          "", // TODO: set url
			"content_type": "json",
			"secret":       "", // TODO: set secret
			"insecure_ssl": "0",
		},
	})
	if err != nil {
		return errors.New("failed to create test webhook", errors.Meta("repo", repoName), err)
	}
	return nil
}

func (c *GitHubTestClient) CreateFile(ctx context.Context, repoName string, branch, path, message string) error {
	_, _, err := c.client.Repositories.CreateFile(ctx, c.Owner, repoName, path, &github.RepositoryContentFileOptions{
		Message: &message,
		Content: []byte(util.RandomHash(32)),
		Branch:  &branch,
	})
	if err != nil {
		return errors.New("failed to create file", errors.Meta("repo", repoName), err)
	}
	return nil
}

func (c *GitHubTestClient) UpdateFile(ctx context.Context, repoName string, branch, path, message string) error {
	_, _, err := c.client.Repositories.UpdateFile(ctx, c.Owner, repoName, path, &github.RepositoryContentFileOptions{
		Message: &message,
		Content: []byte(util.RandomHash(32)),
		Branch:  &branch,
	})
	if err != nil {
		return errors.New("failed to update file", errors.Meta("repo", repoName), err)
	}
	return nil
}

func (c *GitHubTestClient) DeleteFile(ctx context.Context, repoName string, branch, path, message string) error {
	_, _, err := c.client.Repositories.DeleteFile(ctx, c.Owner, repoName, path, &github.RepositoryContentFileOptions{
		Message: &message,
		Content: []byte(util.RandomHash(32)),
		Branch:  &branch,
	})
	if err != nil {
		return errors.New("failed to delete file", errors.Meta("repo", repoName), err)
	}
	return nil
}
