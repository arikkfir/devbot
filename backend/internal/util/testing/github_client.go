package testing

import (
	"context"
	"embed"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "github.com/arikkfir/justest"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v56/github"

	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
)

const (
	GitHubOwner = "devbot-testing"
)

type GClient struct {
	Token  string
	ctx    context.Context
	client *github.Client
}

func GH(ctx context.Context, t T) *GClient {
	token := os.Getenv("GITHUB_TOKEN")
	With(t).Verify(token).Will(Not(BeEmpty())).OrFail()

	c := github.NewClient(nil).WithAuthToken(token)
	req, err := c.NewRequest("GET", "user", nil)
	With(t).Verify(err).Will(BeNil()).OrFail()

	response, err := c.Do(ctx, req, nil)
	With(t).Verify(err).Will(BeNil()).OrFail()
	With(t).Verify(response.StatusCode).Will(EqualTo(http.StatusOK)).OrFail()

	return &GClient{Token: token, ctx: ctx, client: c}
}

func (gh *GClient) CreateRepository(t T, fs embed.FS, embeddedPath string) *GitHubRepositoryInfo {
	// Create the repository
	ghRepo, _, err := gh.client.Repositories.Create(gh.ctx, GitHubOwner, &github.Repository{
		Name:          &[]string{stringsutil.Name()}[0],
		DefaultBranch: github.String("main"),
		Visibility:    &[]string{"public"}[0],
	})
	With(t).Verify(err).Will(BeNil()).OrFail()

	cloneURL := ghRepo.GetCloneURL()
	repoOwner := ghRepo.Owner.GetLogin()
	repoName := ghRepo.GetName()
	t.Cleanup(func() {
		With(t).Verify(gh.client.Repositories.Delete(gh.ctx, repoOwner, repoName)).Will(Succeed()).OrFail()
	})

	// Create the repository contents locally
	// Corresponds to: git init
	path := filepath.Join(os.TempDir(), stringsutil.RandomHash(7))
	localRepo, err := git.PlainInit(path, false)
	With(t).Verify(err).Will(BeNil()).OrFail()
	t.Cleanup(func() { With(t).Verify(os.RemoveAll(path)).Will(Succeed()).OrFail() })

	// Populate the new local repository & commit the changes to HEAD
	worktree, err := localRepo.Worktree()
	With(t).Verify(err).Will(BeNil()).OrFail()
	With(t).Verify(TraverseEmbeddedPath(fs, embeddedPath, func(p string, data []byte) error {
		p = strings.TrimPrefix(p, embeddedPath+"/")
		f := filepath.Join(path, p)
		dir := filepath.Dir(f)
		With(t).Verify(os.MkdirAll(dir, 0755)).Will(Succeed()).OrFail()
		With(t).Verify(os.WriteFile(f, data, 0644)).Will(Succeed()).OrFail()
		With(t).Verify(worktree.Add(p)).Will(Succeed()).OrFail()
		return nil
	})).Will(Succeed()).OrFail()
	With(t).Verify(worktree.Commit("Initial commit", &git.CommitOptions{})).Will(Succeed()).OrFail()

	// Rename local HEAD to "main"
	// Corresponds to:
	// - git branch -m main
	// - git remote add origin https://github.com/devbot-testing/REPOSITORY_NAME.git
	headRef, err := localRepo.Head()
	With(t).Verify(err).Will(BeNil()).OrFail()

	mainRef := plumbing.NewHashReference("refs/heads/main", headRef.Hash())
	With(t).Verify(localRepo.Storer.SetReference(mainRef)).Will(Succeed()).OrFail()
	With(t).Verify(localRepo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{cloneURL}})).Will(Succeed()).OrFail()

	// Push changes to the repository
	// Corresponds to: git push -u origin main
	With(t).Verify(localRepo.PushContext(gh.ctx, &git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/main:refs/heads/main"},
		Progress:   os.Stderr,
		Auth:       &githttp.BasicAuth{Username: "anything", Password: gh.Token},
	})).Will(Succeed()).OrFail()

	repoInfo := &GitHubRepositoryInfo{
		Owner: repoOwner,
		Name:  repoName,
		gh:    gh,
	}

	return repoInfo
}
