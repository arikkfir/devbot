package e2e

import (
	"context"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	GitHubOwner = "devbot-testing"
)

var (
	smeeTunnelRE = regexp.MustCompile("\nConnected (https://smee.io/[a-zA-Z0-9]+)\n")
)

func CreateGitHubClient(ctx context.Context, gh **github.Client) {
	token := os.Getenv("GITHUB_TOKEN")
	Expect(token).ToNot(BeEmpty())

	ghc := github.NewClient(nil).WithAuthToken(token)
	req, err := ghc.NewRequest("GET", "user", nil)
	Expect(err).ToNot(HaveOccurred())
	Expect(req).ToNot(BeNil())
	_, err = ghc.Do(ctx, req, nil)
	Expect(err).ToNot(HaveOccurred())

	*gh = ghc
	DeferCleanup(func() { *gh = nil })
}

func CreateGitHubRepository(ctx context.Context, gh *github.Client, owner *string, repoName *string) {
	*owner = GitHubOwner
	*repoName = stringsutil.Name()
	Expect(gh.Repositories.Create(ctx, GitHubOwner, &github.Repository{
		Name:       &[]string{*repoName}[0],
		Visibility: &[]string{"public"}[0],
	})).Error().ShouldNot(HaveOccurred())

	DeferCleanup(func() {
		_, err := gh.Repositories.Delete(context.Background(), GitHubOwner, *repoName)
		Expect(err).NotTo(HaveOccurred())
	})

	_, _, err := gh.Repositories.CreateFile(ctx, GitHubOwner, *repoName, "README.md", &github.RepositoryContentFileOptions{
		Message: &[]string{"Initial commit"}[0],
		Content: []byte(stringsutil.RandomHash(32)),
		Branch:  &[]string{"main"}[0],
	})
	Expect(err).NotTo(HaveOccurred())

	// Wait until GitHub acknowledges the commit
	Eventually(func(o Gomega) string {
		branchRef, _, err := gh.Git.GetRef(ctx, GitHubOwner, *repoName, "heads/main")
		o.Expect(err).ToNot(HaveOccurred())
		o.Expect(branchRef).ToNot(BeNil())
		return branchRef.GetObject().GetSHA()
	}, 30*time.Second).Should(Not(BeEmpty()))
}

func CreateGitHubRepositoryWithWebhook(ctx context.Context, gh *github.Client, owner *string, repoName *string, webhookSecret *string) {
	CreateGitHubRepository(ctx, gh, owner, repoName)

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
	_, _, err := (*gh).Repositories.CreateHook(ctx, GitHubOwner, *repoName, &github.Hook{
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

func CreateGitHubBranch(ctx context.Context, gh *github.Client, owner, repoName, branch string) {
	mainRef, _, err := gh.Git.GetRef(ctx, owner, repoName, "heads/main")
	Expect(err).ToNot(HaveOccurred())
	Expect(mainRef).ToNot(BeNil())

	branchRef, _, err := gh.Git.CreateRef(ctx, owner, repoName, &github.Reference{
		Ref:    &[]string{"refs/heads/" + branch}[0],
		Object: mainRef.Object,
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(branchRef).ToNot(BeNil())
}

func GetGitHubBranchCommitSHA(ctx context.Context, gh *github.Client, owner, repoName, branch string, sha *string) {
	Eventually(func(o Gomega) string {
		branchRef, _, err := gh.Git.GetRef(ctx, owner, repoName, "heads/"+branch)
		o.Expect(err).ToNot(HaveOccurred())
		o.Expect(branchRef).ToNot(BeNil())
		*sha = branchRef.GetObject().GetSHA()
		return branchRef.GetObject().GetSHA()
	}, 30*time.Second).Should(Not(BeEmpty()))
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
