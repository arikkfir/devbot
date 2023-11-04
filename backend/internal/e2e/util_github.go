package e2e

import (
	"bytes"
	"context"
	"github.com/arikkfir/devbot/backend/internal/util"
	"github.com/google/go-github/v56/github"
	"github.com/secureworks/errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

var (
	tunnelRE = regexp.MustCompile("\nConnected (https://smee.io/[a-zA-Z0-9]+)\n")
)

type GitHubTestClient struct {
	Owner          string
	WebhooksSecret string
	t              *testing.T
	client         *github.Client
	cleanup        []func() error
}

func NewGitHubTestClient(t *testing.T, owner string) *GitHubTestClient {
	t.Helper()
	token := os.Getenv("GITHUB_TOKEN")
	gh := &GitHubTestClient{
		Owner:          owner,
		t:              t,
		client:         github.NewClient(nil).WithAuthToken(token),
		WebhooksSecret: os.Getenv("WEBHOOK_SECRET"),
	}
	t.Cleanup(gh.Close)
	return gh
}

func (c *GitHubTestClient) Close() {
	c.t.Helper()
	for _, f := range c.cleanup {
		if err := f(); err != nil {
			c.t.Errorf("GitHub cleanup function failed: %+v", err)
		}
	}
}

func (c *GitHubTestClient) Cleanup(f func() error) {
	c.t.Helper()
	c.cleanup = append([]func() error{f}, c.cleanup...)
}

func (c *GitHubTestClient) CreateRepository(ctx context.Context) (*github.Repository, error) {
	c.t.Helper()
	repoName := util.RandomHash(7)
	c.t.Logf("Creating GitHub repository '%s/%s'...", c.Owner, repoName)
	repo, _, err := c.client.Repositories.Create(ctx, c.Owner, &github.Repository{
		Name:        &[]string{repoName}[0],
		Description: &[]string{"This is a test repository created by the devbot setup script"}[0],
		Topics:      []string{"devbot", "test"},
		Visibility:  &[]string{"public"}[0],
	})
	if err != nil {
		return nil, errors.New("failed to create test repository", err)
	}
	c.Cleanup(func() error {
		c.t.Helper()
		c.t.Logf("Deleting GitHub repository '%s/%s'...", c.Owner, *repo.Name)
		_, err := c.client.Repositories.Delete(ctx, c.Owner, *repo.Name)
		if err != nil {
			return errors.New("failed to delete test repository", errors.Meta("repo", *repo.FullName), err)
		}
		return nil
	})
	return repo, nil
}

func (c *GitHubTestClient) CreateRepositoryWebhook(ctx context.Context, repoName string, events ...string) error {
	c.t.Helper()
	c.t.Logf("Creating webhook for events '%v' in GitHub repository '%s/%s'...", events, c.Owner, repoName)

	stdout := &bytes.Buffer{}
	cmd := exec.Command("smee", "--port", "8080", "--path", "/github/webhook")
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	c.t.Logf("Starting smee tunnel for webhook of GitHub repository '%s/%s'...", c.Owner, repoName)
	if err := cmd.Start(); err != nil {
		return errors.New("failed to start smee", err)
	}
	c.Cleanup(func() error {
		c.t.Helper()
		c.t.Logf("Closing Smee tunnel for webhook of repository '%s/%s'...", c.Owner, repoName)
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			return errors.New("failed to interrupt smee", err)
		}
		return nil
	})

	var webhookURL string
	for {
		time.Sleep(1 * time.Second)
		if result := tunnelRE.FindStringSubmatch(stdout.String()); len(result) == 2 {
			webhookURL = strings.TrimSpace(result[1])
			break
		} else if cmd.ProcessState != nil {
			c.t.Log(stdout.String())
			return errors.New("failed to find smee tunnel URL, and smee exited")
		}
	}

	_, _, err := c.client.Repositories.CreateHook(ctx, c.Owner, repoName, &github.Hook{
		Name:   &[]string{"web"}[0],
		Active: &[]bool{true}[0],
		Events: events,
		Config: map[string]interface{}{
			"url":          webhookURL,
			"content_type": "json",
			"secret":       c.WebhooksSecret,
			"insecure_ssl": "0",
		},
	})
	if err != nil {
		return errors.New("failed to create test webhook", errors.Meta("repo", repoName), err)
	}

	return nil
}

func (c *GitHubTestClient) CreateFile(ctx context.Context, repoName, branch, path, message string) (*github.RepositoryContentResponse, error) {
	c.t.Helper()
	c.t.Logf("Creating file '%s/%s' in GitHub repository '%s/%s'...", branch, path, c.Owner, repoName)
	cr, _, err := c.client.Repositories.CreateFile(ctx, c.Owner, repoName, path, &github.RepositoryContentFileOptions{
		Message: &message,
		Content: []byte(util.RandomHash(32)),
		Branch:  &branch,
	})
	if err != nil {
		return nil, errors.New("failed to create file", errors.Meta("repo", repoName), err)
	}
	return cr, nil
}

func (c *GitHubTestClient) UpdateFile(ctx context.Context, repoName string, branch, path, message string) error {
	c.t.Helper()
	c.t.Logf("Updating file '%s/%s' in GitHub repository '%s/%s'...", branch, path, c.Owner, repoName)
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
	c.t.Helper()
	c.t.Logf("Deleting file '%s/%s' in GitHub repository '%s/%s'...", branch, path, c.Owner, repoName)
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
