package testing

import (
	"embed"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v56/github"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	GitHubOwner = "devbot-testing"
	gClientKey  = "___github"
)

type GClient struct {
	Token  string
	client *github.Client
}

func GH(t T) *GClient {
	if v := For(t).Value(gClientKey); v == nil {
		token := os.Getenv("GITHUB_TOKEN")
		For(t).Expect(token).Will(Not(BeEmpty()))

		c := github.NewClient(nil).WithAuthToken(token)

		req, err := c.NewRequest("GET", "user", nil)
		For(t).Expect(err).Will(BeNil())

		response, err := c.Do(For(t).Context(), req, nil)
		For(t).Expect(err).Will(BeNil())
		For(t).Expect(response.StatusCode).Will(BeEqualTo(http.StatusOK))
		For(t).AddValue(gClientKey, &GClient{Token: token, client: c})
		return GH(t)
	} else {
		return v.(*GClient)
	}
}

func (gh *GClient) CreateRepository(t T, fs embed.FS, embeddedPath string) *GitHubRepositoryInfo {
	// Create the repository
	ghRepo, _, err := gh.client.Repositories.Create(For(t).Context(), GitHubOwner, &github.Repository{
		Name:          &[]string{stringsutil.Name()}[0],
		DefaultBranch: github.String("main"),
		Visibility:    &[]string{"public"}[0],
	})
	For(t).Expect(err).Will(BeNil())

	cloneURL := ghRepo.GetCloneURL()
	repoOwner := ghRepo.Owner.GetLogin()
	repoName := ghRepo.GetName()
	t.Cleanup(func() {
		For(t).Expect(gh.client.Repositories.Delete(For(t).Context(), repoOwner, repoName)).Will(Succeed())
	})

	// Create the repository contents locally
	// Corresponds to: git init
	path := filepath.Join(os.TempDir(), stringsutil.RandomHash(7))
	localRepo, err := git.PlainInit(path, false)
	For(t).Expect(err).Will(BeNil())
	t.Cleanup(func() { For(t).Expect(os.RemoveAll(path)).Will(Succeed()) })

	// Populate the new local repository & commit the changes to HEAD
	worktree, err := localRepo.Worktree()
	For(t).Expect(err).Will(BeNil())
	For(t).Expect(TraverseEmbeddedPath(fs, embeddedPath, func(p string, data []byte) error {
		p = strings.TrimPrefix(p, embeddedPath+"/")
		f := filepath.Join(path, p)
		dir := filepath.Dir(f)
		For(t).Expect(os.MkdirAll(dir, 0755)).Will(Succeed())
		For(t).Expect(os.WriteFile(f, data, 0644)).Will(Succeed())
		For(t).Expect(worktree.Add(p)).Will(Succeed())
		return nil
	})).Will(Succeed())
	For(t).Expect(worktree.Commit("Initial commit", &git.CommitOptions{})).Will(Succeed())

	// Rename local HEAD to "main"
	// Corresponds to:
	// - git branch -m main
	// - git remote add origin https://github.com/devbot-testing/REPOSITORY_NAME.git
	headRef, err := localRepo.Head()
	For(t).Expect(err).Will(BeNil())

	mainRef := plumbing.NewHashReference("refs/heads/main", headRef.Hash())
	For(t).Expect(localRepo.Storer.SetReference(mainRef)).Will(Succeed())
	For(t).Expect(localRepo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{cloneURL}})).Will(Succeed())

	// Push changes to the repository
	// Corresponds to: git push -u origin main
	For(t).Expect(localRepo.PushContext(For(t).Context(), &git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/main:refs/heads/main"},
		Progress:   os.Stderr,
		Auth:       &githttp.BasicAuth{Username: "anything", Password: gh.Token},
	})).Will(Succeed())

	repoInfo := &GitHubRepositoryInfo{
		Owner: repoOwner,
		Name:  repoName,
		gh:    gh,
	}

	return repoInfo
}
