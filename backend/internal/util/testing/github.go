package testing

import (
	"context"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gbytes"
	corev1 "k8s.io/api/core/v1"
	"os"
	"os/exec"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

const (
	GitHubOwner           = "devbot-testing"
	GitHubSecretName      = "devbot-github-auth"
	GitHubSecretNamespace = "devbot"
	GitHubSecretPATKey    = "TOKEN"
)

var (
	smeeTunnelRE = regexp.MustCompile("\nConnected (https://smee.io/[a-zA-Z0-9]+)\n")
)

func CreateGitHubClient(ctx context.Context, k8s client.Client, Client **github.Client) {
	gitHubSecretKey := client.ObjectKey{Namespace: GitHubSecretNamespace, Name: GitHubSecretName}
	secret := &corev1.Secret{}
	Expect(k8s.Get(ctx, gitHubSecretKey, secret)).To(Succeed())
	*Client = github.NewClient(nil).WithAuthToken(ObtainGitHubPAT(ctx, k8s))

	DeferCleanup(func() { *Client = nil })
}

func CreateGitHubRepository(ctx context.Context, k8s client.Client, gh *github.Client, repoName *string) {
	*repoName = stringsutil.Name()
	Expect(gh.Repositories.Create(ctx, GitHubOwner, &github.Repository{
		Name:       &[]string{*repoName}[0],
		Visibility: &[]string{"public"}[0],
	})).Error().ShouldNot(HaveOccurred())

	var smeeCommand *exec.Cmd
	smeeOutput := gbytes.NewBuffer()
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
			"secret":       ObtainGitHubWebhookSecret(ctx, k8s),
			"insecure_ssl": "0",
		},
	})
	Expect(err).NotTo(HaveOccurred())

	_, _, err = gh.Repositories.CreateFile(ctx, GitHubOwner, *repoName, "README.md", &github.RepositoryContentFileOptions{
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

	DeferCleanup(func() {
		_, err := gh.Repositories.Delete(context.Background(), GitHubOwner, *repoName)
		Expect(err).NotTo(HaveOccurred())
	})
}

func CreateGitHubBranch(ctx context.Context, gh *github.Client, repoName, branch string, deleteOnCleanup bool) {
	mainRef, _, err := gh.Git.GetRef(ctx, GitHubOwner, repoName, "heads/main")
	Expect(err).ToNot(HaveOccurred())
	Expect(mainRef).ToNot(BeNil())

	branchRef, _, err := gh.Git.CreateRef(ctx, GitHubOwner, repoName, &github.Reference{
		Ref:    &[]string{"refs/heads/" + branch}[0],
		Object: mainRef.Object,
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(branchRef).ToNot(BeNil())

	if deleteOnCleanup {
		DeferCleanup(func() {
			_, err := gh.Git.DeleteRef(context.Background(), GitHubOwner, repoName, "heads/"+branch)
			Expect(err).NotTo(HaveOccurred())
		})
	}
}

func DeleteGitHubBranch(ctx context.Context, gh *github.Client, repoName, branch string) {
	_, err := gh.Git.DeleteRef(ctx, GitHubOwner, repoName, "heads/"+branch)
	Expect(err).ToNot(HaveOccurred())
}

func GetGitHubBranchCommitSHA(ctx context.Context, gh *github.Client, repoName, branch string, sha *string) {
	Eventually(func(o Gomega) string {
		branchRef, _, err := gh.Git.GetRef(ctx, GitHubOwner, repoName, "heads/"+branch)
		o.Expect(err).ToNot(HaveOccurred())
		o.Expect(branchRef).ToNot(BeNil())
		*sha = branchRef.GetObject().GetSHA()
		return branchRef.GetObject().GetSHA()
	}, 2*time.Minute).Should(Not(BeEmpty()))
	DeferCleanup(func() { *sha = "" })
}

func CreateGitHubFile(ctx context.Context, gh *github.Client, repoName, branch string, sha *string) {
	branchRef, _, err := gh.Repositories.GetBranch(ctx, GitHubOwner, repoName, branch, 0)
	Expect(err).ToNot(HaveOccurred())
	Expect(branchRef).ToNot(BeNil())

	cr, _, err := gh.Repositories.CreateFile(ctx, GitHubOwner, repoName, stringsutil.RandomHash(7)+".txt", &github.RepositoryContentFileOptions{
		Message: &[]string{stringsutil.RandomHash(32)}[0],
		Content: []byte(stringsutil.RandomHash(32)),
		Branch:  &branch,
	})
	Expect(err).ToNot(HaveOccurred())

	*sha = cr.GetSHA()
}
