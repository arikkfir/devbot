package util

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	stringsutil "github.com/arikkfir/devbot/internal/util/strings"
)

const (
	GitHubOwner = "devbot-testing"
)

func GetGitHubToken() string {
	GinkgoHelper()
	token := os.Getenv("GITHUB_TOKEN")
	Expect(token).NotTo(BeEmpty())
	return token
}

func NewGitHubClient(ctx context.Context) *github.Client {
	GinkgoHelper()
	c := github.NewClient(nil).WithAuthToken(GetGitHubToken())
	req, err := c.NewRequest("GET", "user", nil)
	Expect(err).To(BeNil())

	response, err := c.Do(ctx, req, nil)
	Expect(err).To(BeNil())
	Expect(response.StatusCode).To(Equal(http.StatusOK))
	return c
}

func CreateGitHubRepository(ctx context.Context, ghc *github.Client, fs embed.FS, embeddedPath string) *github.Repository {
	GinkgoHelper()

	// Create the repository
	ghRepo, _, err := ghc.Repositories.Create(ctx, GitHubOwner, &github.Repository{
		Name:          &[]string{stringsutil.Name()}[0],
		DefaultBranch: github.String("main"),
		Visibility:    &[]string{"public"}[0],
	})
	Expect(err).To(BeNil())
	Eventually(func(g Gomega) {
		// Sometimes it takes a short while for the repository to get created - thus we use Eventually to wait for it
		Expect(ghc.Repositories.Get(ctx, ghRepo.Owner.GetLogin(), ghRepo.GetName())).Error().Error().To(Succeed())
	}, "30s")

	// Create the repository contents locally
	// Corresponds to: git init
	path := filepath.Join(os.TempDir(), stringsutil.RandomHash(7))
	localRepo, err := git.PlainInit(path, false)
	Expect(err).To(BeNil())
	DeferCleanup(func() { Expect(os.RemoveAll(path)).To(Succeed()) })

	// Populate the new local repository & commit the changes to HEAD
	worktree, err := localRepo.Worktree()
	Expect(err).To(BeNil())
	Expect(TraverseEmbeddedPath(fs, embeddedPath, func(p string, data []byte) error {
		p = strings.TrimPrefix(p, embeddedPath+"/")
		f := filepath.Join(path, p)
		dir := filepath.Dir(f)
		Expect(os.MkdirAll(dir, 0755)).To(Succeed())
		Expect(os.WriteFile(f, data, 0644)).To(Succeed())
		Expect(worktree.Add(p)).Error().To(Succeed())
		return nil
	})).To(Succeed())
	Expect(worktree.Commit("Initial commit", &git.CommitOptions{
		Author:    &object.Signature{Name: "CI", Email: "arik@kfirs.com", When: time.Now()},
		Committer: &object.Signature{Name: "CI", Email: "arik@kfirs.com", When: time.Now()},
	})).Error().To(BeNil())

	// Rename local HEAD to "main"
	// Corresponds to:
	// - git branch -m main
	// - git remote add origin https://github.com/devbot-testing/REPOSITORY_NAME.git
	headRef, err := localRepo.Head()
	Expect(err).To(BeNil())

	mainRef := plumbing.NewHashReference("refs/heads/main", headRef.Hash())
	Expect(localRepo.Storer.SetReference(mainRef)).To(Succeed())
	Expect(localRepo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{ghRepo.GetCloneURL()}})).Error().To(Succeed())

	// Push changes to the repository
	// Corresponds to: git push -u origin main
	Expect(localRepo.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/main:refs/heads/main"},
		Progress:   os.Stderr,
		Auth:       &githttp.BasicAuth{Username: "anything", Password: GetGitHubToken()},
	})).To(Succeed())

	DeferCleanup(func(ctx context.Context) {
		Expect(ghc.Repositories.Delete(ctx, ghRepo.Owner.GetLogin(), ghRepo.GetName())).Error().To(Succeed())
	})

	return ghRepo
}

func CloneGitHubRepository(ctx context.Context, ghRepo *github.Repository) *git.Repository {
	GinkgoHelper()
	clonePath := filepath.Join(os.TempDir(), stringsutil.RandomHash(7))
	DeferCleanup(func() { Expect(os.RemoveAll(clonePath)).To(Succeed()) })

	cloneRepo, err := git.PlainCloneContext(ctx, clonePath, false, &git.CloneOptions{
		ReferenceName: "refs/heads/main",
		URL:           *ghRepo.CloneURL,
	})
	Expect(err).To(BeNil())
	return cloneRepo
}

func CreateGitHubWebhookTunnel() string {
	GinkgoHelper()
	var smeeCommand *exec.Cmd
	smeeOutput := bytes.Buffer{}
	smeeCommand = exec.Command("smee", "--port", "8080", "--path", "/github/webhook")
	smeeCommand.Stdout = &smeeOutput
	smeeCommand.Stderr = os.Stderr
	Expect(smeeCommand.Start()).To(Succeed())
	DeferCleanup(func() { Expect(smeeCommand.Process.Signal(os.Interrupt)).To(Succeed()) })
	Eventually(smeeCommand.Stdout, "10s").Should(Say("\nConnected "))

	webHookURL, err := smeeOutput.ReadString('\n')
	Expect(err).To(BeNil())
	return webHookURL
}

func CreateGitHubRepositoryBranch(ctx context.Context, ghc *github.Client, ghRepo *github.Repository, branch string) {
	GinkgoHelper()
	mainRef, _, err := ghc.Git.GetRef(ctx, ghRepo.Owner.GetLogin(), ghRepo.GetName(), "heads/main")
	Expect(err).To(BeNil())
	Expect(mainRef).ToNot(BeNil())
	Expect(ghc.Git.CreateRef(ctx, ghRepo.Owner.GetLogin(), ghRepo.GetName(), &github.Reference{
		Ref:    github.String("refs/heads/" + branch),
		Object: mainRef.Object,
	})).Error().To(BeNil())
}

func GitHubRepositoryBranchExists(ctx context.Context, ghc *github.Client, ghRepo *github.Repository, branch string) bool {
	GinkgoHelper()
	_, resp, err := ghc.Repositories.GetBranch(ctx, ghRepo.Owner.GetLogin(), ghRepo.GetName(), branch, 0)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return false
		} else {
			Fail(fmt.Sprintf("Failed to get branch '%s' from repository '%s/%s': %s", branch, ghRepo.Owner.GetLogin(), ghRepo.GetName(), err))
		}
	}
	return true
}

func GetGitHubRepositoryBranchSHA(ctx context.Context, ghc *github.Client, ghRepo *github.Repository, branch string) string {
	GinkgoHelper()
	var sha string
	branchRef, _, err := ghc.Git.GetRef(ctx, ghRepo.Owner.GetLogin(), ghRepo.GetName(), "heads/"+branch)
	Expect(err).To(BeNil())
	Expect(branchRef).ToNot(BeNil())
	sha = branchRef.GetObject().GetSHA()
	Expect(sha).ToNot(BeEmpty())
	return sha
}

func CreateFileInGitHubRepositoryBranch(ctx context.Context, ghc *github.Client, ghRepo *github.Repository, branch string) string {
	GinkgoHelper()
	var sha string
	file := stringsutil.RandomHash(7) + ".txt"
	cr, _, err := ghc.Repositories.CreateFile(ctx, ghRepo.Owner.GetLogin(), ghRepo.GetName(), file, &github.RepositoryContentFileOptions{
		Message: github.String(stringsutil.RandomHash(32)),
		Content: []byte(stringsutil.RandomHash(32)),
		Branch:  &branch,
	})
	Expect(err).To(BeNil())
	sha = cr.GetSHA()
	Expect(sha).ToNot(BeEmpty())
	return sha
}

func DeleteGitHubRepositoryBranch(ctx context.Context, ghc *github.Client, ghRepo *github.Repository, branch string) {
	GinkgoHelper()
	Expect(ghc.Git.DeleteRef(ctx, ghRepo.Owner.GetLogin(), ghRepo.GetName(), "heads/"+branch)).To(Succeed())
}

func GetGitHubRepositoryBranchNamesAndSHA(ctx context.Context, ghc *github.Client, ghRepo *github.Repository) map[string]string {
	GinkgoHelper()
	result := make(map[string]string)
	branchesListOptions := &github.BranchListOptions{}
	for {
		branchesList, response, err := ghc.Repositories.ListBranches(ctx, ghRepo.Owner.GetLogin(), ghRepo.GetName(), branchesListOptions)
		Expect(err).To(BeNil())
		for _, branch := range branchesList {
			branchName := branch.GetName()
			revision := branch.GetCommit().GetSHA()
			result[branchName] = revision
		}
		if response.NextPage == 0 {
			break
		}
		branchesListOptions.Page = response.NextPage
	}
	return result
}
