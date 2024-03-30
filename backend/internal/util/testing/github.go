package testing

import (
	"bytes"
	"context"
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
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	GitHubOwner    = "devbot-testing"
	TenSecs        = 10 * time.Second
	HundredsMillis = 100 * time.Millisecond
)

var (
	smeeTunnelRE = regexp.MustCompile("\nConnected (https://smee.io/[a-zA-Z0-9]+)\n")
)

type GitHub struct {
	Token        string
	Repositories map[string]*GitHubRepositoryInfo
	client       *github.Client
}

func NewGitHub(t JustT, ctx context.Context) *GitHub {
	token := os.Getenv("GITHUB_TOKEN")
	For(t).Expect(token).WillNot(BeEmpty())

	c := github.NewClient(nil).WithAuthToken(token)

	req, err := c.NewRequest("GET", "user", nil)
	For(t).Expect(err).Will(BeNil())

	response, err := c.Do(ctx, req, nil)
	For(t).Expect(err).Will(BeNil())
	For(t).Expect(response.StatusCode).Will(BeEqualTo(http.StatusOK))

	return &GitHub{
		Token:        token,
		Repositories: make(map[string]*GitHubRepositoryInfo),
		client:       c,
	}
}

func (gh *GitHub) CreateRepository(t JustT, ctx context.Context, fs embed.FS, embeddedPath string) *GitHubRepositoryInfo {
	// Create the repository
	ghRepo, _, err := gh.client.Repositories.Create(ctx, GitHubOwner, &github.Repository{
		Name:          &[]string{stringsutil.Name()}[0],
		DefaultBranch: github.String("main"),
		Visibility:    &[]string{"public"}[0],
	})
	For(t).Expect(err).Will(BeNil())

	cloneURL := ghRepo.GetCloneURL()
	repoOwner := ghRepo.Owner.GetLogin()
	repoName := ghRepo.GetName()
	t.Cleanup(func() { For(t).Expect(gh.client.Repositories.Delete(ctx, repoOwner, repoName)).Will(Succeed()) })

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
	For(t).Expect(localRepo.PushContext(ctx, &git.PushOptions{
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
	gh.Repositories[ghRepo.GetFullName()] = repoInfo
	t.Cleanup(func() { delete(gh.Repositories, ghRepo.GetFullName()) })

	return repoInfo
}

type GitHubRepositoryInfo struct {
	Owner         string
	Name          string
	WebhookSecret string
	gh            *GitHub
}

func (r *GitHubRepositoryInfo) SetupWebhook(t JustT, ctx context.Context) {
	if r.WebhookSecret != "" {
		return
	}

	webhookSecret := stringsutil.RandomHash(16)

	var smeeCommand *exec.Cmd
	smeeOutput := bytes.Buffer{}
	smeeCommand = exec.Command("smee", "--port", "8080", "--path", "/github/webhook")
	smeeCommand.Stdout = &smeeOutput
	smeeCommand.Stderr = os.Stderr
	For(t).Expect(smeeCommand.Start()).Will(BeNil())
	t.Cleanup(func() { For(t).Expect(smeeCommand.Process.Signal(os.Interrupt)).Will(Succeed()) })

	For(t).Expect(smeeCommand.Stdout).Will(Eventually(Say(smeeTunnelRE)).Within(TenSecs).ProbingEvery(HundredsMillis))
	webHookURL := strings.TrimSpace(smeeTunnelRE.FindStringSubmatch(smeeOutput.String())[1])

	hook, _, err := r.gh.client.Repositories.CreateHook(ctx, r.Owner, r.Name, &github.Hook{
		Name:   github.String("web"),
		Active: github.Bool(true),
		Events: []string{"push"},
		Config: map[string]interface{}{
			"url":          webHookURL,
			"content_type": "json",
			"secret":       webhookSecret,
			"insecure_ssl": "0",
		},
	})
	For(t).Expect(err).Will(BeNil())
	t.Cleanup(func() {
		For(t).Expect(r.gh.client.Repositories.DeleteHook(ctx, r.Owner, r.Name, hook.GetID())).Will(Succeed())
	})

	r.WebhookSecret = webhookSecret
}

func (r *GitHubRepositoryInfo) CreateBranch(t JustT, ctx context.Context, branch string) string {
	mainRef, _, err := r.gh.client.Git.GetRef(ctx, r.Owner, r.Name, "heads/main")
	For(t).Expect(err).Will(BeNil())
	For(t).Expect(mainRef).WillNot(BeNil())
	For(t).Expect(r.gh.client.Git.CreateRef(ctx, r.Owner, r.Name, &github.Reference{
		Ref:    github.String("refs/heads/" + branch),
		Object: mainRef.Object,
	})).Will(Succeed())
	return r.GetBranchSHA(t, ctx, branch)
}

func (r *GitHubRepositoryInfo) GetBranchSHA(t JustT, ctx context.Context, branch string) string {
	var sha string
	For(t).Expect(func(t JustT) {
		branchRef, _, err := r.gh.client.Git.GetRef(ctx, r.Owner, r.Name, "heads/"+branch)
		For(t).Expect(err).Will(BeNil())
		For(t).Expect(branchRef).WillNot(BeNil())
		sha = branchRef.GetObject().GetSHA()
		For(t).Expect(sha).WillNot(BeEmpty())
	}).Will(Eventually(Succeed()).Within(30 * time.Second).ProbingEvery(time.Second))

	return sha
}

func (r *GitHubRepositoryInfo) CreateFile(t JustT, ctx context.Context, branch string) string {
	var sha string
	For(t).Expect(func(t JustT) {
		branchRef, _, err := r.gh.client.Repositories.GetBranch(ctx, r.Owner, r.Name, branch, 0)
		For(t).Expect(err).Will(BeNil())
		For(t).Expect(branchRef).WillNot(BeNil())

		file := stringsutil.RandomHash(7) + ".txt"
		cr, _, err := r.gh.client.Repositories.CreateFile(ctx, r.Owner, r.Name, file, &github.RepositoryContentFileOptions{
			Message: github.String(stringsutil.RandomHash(32)),
			Content: []byte(stringsutil.RandomHash(32)),
			Branch:  &branch,
		})
		For(t).Expect(err).Will(BeNil())

		sha = cr.GetSHA()
		For(t).Expect(sha).WillNot(BeEmpty())
	}).Will(Eventually(Succeed()).Within(30 * time.Second).ProbingEvery(time.Second))
	return sha
}

func (r *GitHubRepositoryInfo) DeleteBranch(t JustT, ctx context.Context, branch string) {
	For(t).Expect(r.gh.client.Git.DeleteRef(ctx, r.Owner, r.Name, "heads/"+branch)).Will(Succeed())
}
