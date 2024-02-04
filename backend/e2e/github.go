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
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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
	k8sServiceAccountKind = reflect.TypeOf(corev1.ServiceAccount{}).Name()
	k8sClusterRoleKind    = reflect.TypeOf(rbacv1.ClusterRole{}).Name()
	smeeTunnelRE          = regexp.MustCompile("\nConnected (https://smee.io/[a-zA-Z0-9]+)\n")
)

func CreateGitHubClient(ctx context.Context, gh **github.Client, gitHubToken *string) {
	token := os.Getenv("GITHUB_TOKEN")
	Expect(token).ToNot(BeEmpty())

	ghc := github.NewClient(nil).WithAuthToken(token)
	req, err := ghc.NewRequest("GET", "user", nil)
	Expect(err).ToNot(HaveOccurred())
	Expect(req).ToNot(BeNil())
	_, err = ghc.Do(ctx, req, nil)
	Expect(err).ToNot(HaveOccurred())

	*gh = ghc
	*gitHubToken = token
	DeferCleanup(func() { *gh = nil })
}

func CreateGitHubRepository(ctx context.Context, gh *github.Client, embeddedPath string, owner, repoName, mainSHA *string) {
	token := os.Getenv("GITHUB_TOKEN")
	Expect(token).ToNot(BeEmpty())

	*owner = GitHubOwner
	*repoName = stringsutil.Name()

	// Create the repository
	ghRepo, _, err := gh.Repositories.Create(ctx, *owner, &github.Repository{
		Name:          &[]string{*repoName}[0],
		DefaultBranch: github.String("main"),
		Visibility:    &[]string{"public"}[0],
	})
	Expect(err).ToNot(HaveOccurred())
	DeferCleanup(func(ctx context.Context) {
		Expect(gh.Repositories.Delete(ctx, *owner, *repoName)).Error().NotTo(HaveOccurred())
	})

	// Create the repository contents locally
	// Corresponds to: git init
	path := filepath.Join(os.TempDir(), stringsutil.RandomHash(7))
	localRepo, err := git.PlainInit(path, false)
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(os.RemoveAll(path)).To(Succeed()) })

	// Populate the new local repository & commit the changes to HEAD
	worktree, err := localRepo.Worktree()
	Expect(err).ToNot(HaveOccurred())
	Expect(traverseEmbeddedPath(embeddedPath, func(p string, data []byte) error {
		f := filepath.Join(path, p)
		dir := filepath.Dir(f)
		Expect(os.MkdirAll(dir, 0755)).To(Succeed())
		Expect(os.WriteFile(f, data, 0644)).To(Succeed())
		Expect(worktree.Add(p)).Error().NotTo(HaveOccurred())
		return nil
	})).To(Succeed())
	Expect(worktree.Commit("Initial commit", &git.CommitOptions{})).Error().NotTo(HaveOccurred())

	// Rename local HEAD to "main"
	// Corresponds to:
	// - git branch -m main
	// - git remote add origin https://github.com/devbot-testing/REPOSITORY_NAME.git
	headRef, err := localRepo.Head()
	Expect(err).ToNot(HaveOccurred())
	Expect(localRepo.Storer.SetReference(plumbing.NewHashReference("refs/heads/main", headRef.Hash()))).To(Succeed())
	Expect(localRepo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{ghRepo.GetCloneURL()}})).Error().NotTo(HaveOccurred())

	// Push changes to the repository
	// Corresponds to: git push -u origin main
	Expect(localRepo.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/main:refs/heads/main"},
		Progress:   GinkgoWriter,
		Auth:       &http.BasicAuth{Username: "anything", Password: token},
	})).To(Succeed())

	// Wait until GitHub acknowledges the branch & commit
	GetGitHubBranchCommitSHA(ctx, gh, *owner, *repoName, "main", mainSHA)
}

func CreateGitHubRepositoryWithWebhook(ctx context.Context, gh *github.Client, embeddedPath string, owner, repoName, mainSHA *string, webhookSecret *string) {
	CreateGitHubRepository(ctx, gh, embeddedPath, owner, repoName, mainSHA)

	*webhookSecret = stringsutil.RandomHash(16)

	var smeeCommand *exec.Cmd
	smeeOutput := NewBuffer()
	smeeCommand = exec.Command("smee", "--port", "8080", "--path", "/github/webhook")
	smeeCommand.Stdout = smeeOutput
	smeeCommand.Stderr = os.Stderr
	Expect(smeeCommand.Start()).To(Succeed())
	DeferCleanup(func() error { return smeeCommand.Process.Signal(os.Interrupt) })

	Eventually(smeeCommand.Stdout).Within(10 * time.Second).Should(Say(smeeTunnelRE.String()))
	webHookURL := strings.TrimSpace(smeeTunnelRE.FindStringSubmatch(string(smeeOutput.Contents()))[1])
	_, _, err := (*gh).Repositories.CreateHook(ctx, *owner, *repoName, &github.Hook{
		Name:   &[]string{"web"}[0],
		Active: &[]bool{true}[0],
		Events: []string{"push"},
		Config: map[string]interface{}{
			"url":          webHookURL,
			"content_type": "json",
			"secret":       *webhookSecret,
			"insecure_ssl": "0",
		},
	})
	Expect(err).NotTo(HaveOccurred())
}

func CreateGitHubBranch(ctx context.Context, gh *github.Client, owner, repoName, branch string, sha *string) {
	mainRef, _, err := gh.Git.GetRef(ctx, owner, repoName, "heads/main")
	Expect(err).ToNot(HaveOccurred())
	Expect(mainRef).ToNot(BeNil())
	Expect(gh.Git.CreateRef(ctx, owner, repoName, &github.Reference{
		Ref:    &[]string{"refs/heads/" + branch}[0],
		Object: mainRef.Object,
	})).Error().NotTo(HaveOccurred())
	GetGitHubBranchCommitSHA(ctx, gh, owner, repoName, branch, sha)
}

func GetGitHubBranchCommitSHA(ctx context.Context, gh *github.Client, owner, repoName, branch string, sha *string) {
	Eventually(func(o Gomega) {
		branchRef, _, err := gh.Git.GetRef(ctx, owner, repoName, "heads/"+branch)
		o.Expect(err).ToNot(HaveOccurred())
		o.Expect(branchRef).ToNot(BeNil())
		if sha != nil {
			*sha = branchRef.GetObject().GetSHA()
			Expect(*sha).ToNot(BeEmpty())
		}
	}, 30*time.Second).Should(Succeed())
	DeferCleanup(func() { *sha = "" })
}

func CreateGitHubFile(ctx context.Context, gh *github.Client, owner, repoName, branch string, sha *string) {
	branchRef, _, err := gh.Repositories.GetBranch(ctx, owner, repoName, branch, 0)
	Expect(err).ToNot(HaveOccurred())
	Expect(branchRef).ToNot(BeNil())

	cr, _, err := gh.Repositories.CreateFile(ctx, owner, repoName, stringsutil.RandomHash(7)+".txt", &github.RepositoryContentFileOptions{
		Message: &[]string{stringsutil.RandomHash(32)}[0],
		Content: []byte(stringsutil.RandomHash(32)),
		Branch:  &branch,
	})
	Expect(err).ToNot(HaveOccurred())

	*sha = cr.GetSHA()
}

func DeleteGitHubBranch(ctx context.Context, gh *github.Client, owner, repoName, branch string) {
	_, err := gh.Git.DeleteRef(ctx, owner, repoName, "heads/"+branch)
	Expect(err).ToNot(HaveOccurred())
}
