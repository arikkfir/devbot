package e2e

import (
	"context"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	GitHubOwner                                        = "devbot-testing"
	DevbotNamespace                                    = "devbot"
	DevbotGitHubRepositoryControllerServiceAccountName = "devbot-github-repository-controller"
	DevbotGitHubRefControllerServiceAccountName        = "devbot-github-ref-controller"
)

var (
	smeeTunnelRE = regexp.MustCompile("\nConnected (https://smee.io/[a-zA-Z0-9]+)\n")
)

type GitHub struct {
	Token        string
	Repositories map[string]*GitHubRepositoryInfo

	client *github.Client
}

func NewGitHub(ctx context.Context) *GitHub {
	token := os.Getenv("GITHUB_TOKEN")
	Expect(token).ToNot(BeEmpty())

	c := github.NewClient(nil).WithAuthToken(token)
	req, err := c.NewRequest("GET", "user", nil)
	Expect(err).NotTo(HaveOccurred())
	Expect(req).ToNot(BeNil())
	Expect(c.Do(ctx, req, nil)).Error().NotTo(HaveOccurred())

	return &GitHub{
		Token:        token,
		Repositories: make(map[string]*GitHubRepositoryInfo),
		client:       c,
	}
}

func (gh *GitHub) CreateRepository(ctx context.Context, embeddedPath string) *GitHubRepositoryInfo {
	// Create the repository
	ghRepo, _, err := gh.client.Repositories.Create(ctx, GitHubOwner, &github.Repository{
		Name:          &[]string{stringsutil.Name()}[0],
		DefaultBranch: github.String("main"),
		Visibility:    &[]string{"public"}[0],
	})
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func(ctx context.Context) {
		Expect(gh.client.Repositories.Delete(ctx, ghRepo.Owner.GetLogin(), ghRepo.GetName())).Error().NotTo(HaveOccurred())
	})

	// Create the repository contents locally
	// Corresponds to: git init
	path := filepath.Join(os.TempDir(), stringsutil.RandomHash(7))
	localRepo, err := git.PlainInit(path, false)
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(os.RemoveAll(path)).Error().NotTo(HaveOccurred()) })

	// Populate the new local repository & commit the changes to HEAD
	worktree, err := localRepo.Worktree()
	Expect(err).NotTo(HaveOccurred())
	Expect(traverseEmbeddedPath(embeddedPath, func(p string, data []byte) error {
		f := filepath.Join(path, p)
		dir := filepath.Dir(f)
		Expect(os.MkdirAll(dir, 0755)).Error().NotTo(HaveOccurred())
		Expect(os.WriteFile(f, data, 0644)).Error().NotTo(HaveOccurred())
		Expect(worktree.Add(p)).Error().NotTo(HaveOccurred())
		return nil
	})).To(Succeed())
	Expect(worktree.Commit("Initial commit", &git.CommitOptions{})).Error().NotTo(HaveOccurred())

	// Rename local HEAD to "main"
	// Corresponds to:
	// - git branch -m main
	// - git remote add origin https://github.com/devbot-testing/REPOSITORY_NAME.git
	headRef, err := localRepo.Head()
	Expect(err).NotTo(HaveOccurred())
	Expect(localRepo.Storer.SetReference(plumbing.NewHashReference("refs/heads/main", headRef.Hash()))).Error().NotTo(HaveOccurred())
	Expect(localRepo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{ghRepo.GetCloneURL()}})).Error().NotTo(HaveOccurred())

	// Push changes to the repository
	// Corresponds to: git push -u origin main
	Expect(localRepo.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/main:refs/heads/main"},
		Progress:   GinkgoWriter,
		Auth:       &http.BasicAuth{Username: "anything", Password: gh.Token},
	})).To(Succeed())

	repoInfo := &GitHubRepositoryInfo{
		Owner: ghRepo.Owner.GetLogin(),
		Name:  ghRepo.GetName(),
		gh:    gh,
	}
	gh.Repositories[ghRepo.GetFullName()] = repoInfo
	DeferCleanup(func() { delete(gh.Repositories, ghRepo.GetFullName()) })

	return repoInfo
}

type GitHubRepositoryInfo struct {
	Owner         string
	Name          string
	WebhookSecret string
	gh            *GitHub
}

func (r *GitHubRepositoryInfo) SetupWebhook(ctx context.Context) {
	if r.WebhookSecret != "" {
		return
	}

	webhookSecret := stringsutil.RandomHash(16)

	var smeeCommand *exec.Cmd
	smeeOutput := NewBuffer()
	smeeCommand = exec.Command("smee", "--port", "8080", "--path", "/github/webhook")
	smeeCommand.Stdout = smeeOutput
	smeeCommand.Stderr = os.Stderr
	Expect(smeeCommand.Start()).Error().NotTo(HaveOccurred())
	DeferCleanup(func() error { return smeeCommand.Process.Signal(os.Interrupt) })

	Eventually(smeeCommand.Stdout).Within(10 * time.Second).Should(Say(smeeTunnelRE.String()))
	webHookURL := strings.TrimSpace(smeeTunnelRE.FindStringSubmatch(string(smeeOutput.Contents()))[1])

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
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func(ctx context.Context) {
		Expect(r.gh.client.Repositories.DeleteHook(ctx, r.Owner, r.Name, hook.GetID())).Error().NotTo(HaveOccurred())
	})

	r.WebhookSecret = webhookSecret
}

func (r *GitHubRepositoryInfo) CreateBranch(ctx context.Context, branch string) string {
	mainRef, _, err := r.gh.client.Git.GetRef(ctx, r.Owner, r.Name, "heads/main")
	Expect(err).NotTo(HaveOccurred())
	Expect(mainRef).ToNot(BeNil())
	Expect(r.gh.client.Git.CreateRef(ctx, r.Owner, r.Name, &github.Reference{
		Ref:    github.String("refs/heads/" + branch),
		Object: mainRef.Object,
	})).Error().NotTo(HaveOccurred())
	return r.GetBranchSHA(ctx, branch)
}

func (r *GitHubRepositoryInfo) GetBranchSHA(ctx context.Context, branch string) string {
	var sha string
	Eventually(func(o Gomega) {
		branchRef, _, err := r.gh.client.Git.GetRef(ctx, r.Owner, r.Name, "heads/"+branch)
		o.Expect(err).NotTo(HaveOccurred())
		o.Expect(branchRef).ToNot(BeNil())
		sha = branchRef.GetObject().GetSHA()
		Expect(sha).ToNot(BeEmpty())
	}, 30*time.Second).Should(Succeed())
	return sha
}

func (r *GitHubRepositoryInfo) CreateFile(ctx context.Context, branch string) string {
	var sha string
	Eventually(func(o Gomega) {
		branchRef, _, err := r.gh.client.Repositories.GetBranch(ctx, r.Owner, r.Name, branch, 0)
		Expect(err).NotTo(HaveOccurred())
		Expect(branchRef).ToNot(BeNil())

		file := stringsutil.RandomHash(7) + ".txt"
		cr, _, err := r.gh.client.Repositories.CreateFile(ctx, r.Owner, r.Name, file, &github.RepositoryContentFileOptions{
			Message: github.String(stringsutil.RandomHash(32)),
			Content: []byte(stringsutil.RandomHash(32)),
			Branch:  &branch,
		})
		Expect(err).NotTo(HaveOccurred())
		sha = cr.GetSHA()
		Expect(sha).ToNot(BeEmpty())
	}, 30*time.Second).Should(Succeed())
	return sha
}

func (r *GitHubRepositoryInfo) DeleteBranch(ctx context.Context, branch string) {
	Expect(r.gh.client.Git.DeleteRef(ctx, r.Owner, r.Name, "heads/"+branch)).Error().NotTo(HaveOccurred())
}
