package testing

import (
	"bytes"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"github.com/google/go-github/v56/github"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	TenSecs        = 10 * time.Second
	HundredsMillis = 100 * time.Millisecond
)

var (
	smeeTunnelRE = regexp.MustCompile("\nConnected (https://smee.io/[a-zA-Z0-9]+)\n")
)

type GitHubRepositoryInfo struct {
	Owner         string
	Name          string
	WebhookSecret string
	gh            *GClient
}

func (r *GitHubRepositoryInfo) SetupWebhook(t T) {
	if r.WebhookSecret != "" {
		return
	}

	webhookSecret := stringsutil.RandomHash(16)

	var smeeCommand *exec.Cmd
	smeeOutput := bytes.Buffer{}
	smeeCommand = exec.Command("smee", "--port", "8080", "--path", "/github/webhook")
	smeeCommand.Stdout = &smeeOutput
	smeeCommand.Stderr = os.Stderr
	For(t).Expect(smeeCommand.Start()).Will(Succeed())
	t.Cleanup(func() { For(t).Expect(smeeCommand.Process.Signal(os.Interrupt)).Will(Succeed()) })

	For(t).Expect(smeeCommand.Stdout).Will(Eventually(Say(smeeTunnelRE)).Within(TenSecs).ProbingEvery(HundredsMillis))
	webHookURL := strings.TrimSpace(smeeTunnelRE.FindStringSubmatch(smeeOutput.String())[1])

	hook, _, err := r.gh.client.Repositories.CreateHook(For(t).Context(), r.Owner, r.Name, &github.Hook{
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
		For(t).Expect(r.gh.client.Repositories.DeleteHook(For(t).Context(), r.Owner, r.Name, hook.GetID())).Will(Succeed())
	})

	r.WebhookSecret = webhookSecret
}

func (r *GitHubRepositoryInfo) CreateBranch(t T, branch string) string {
	mainRef, _, err := r.gh.client.Git.GetRef(For(t).Context(), r.Owner, r.Name, "heads/main")
	For(t).Expect(err).Will(BeNil())
	For(t).Expect(mainRef).Will(Not(BeNil()))
	For(t).Expect(r.gh.client.Git.CreateRef(For(t).Context(), r.Owner, r.Name, &github.Reference{
		Ref:    github.String("refs/heads/" + branch),
		Object: mainRef.Object,
	})).Will(Succeed())
	return r.GetBranchSHA(t, branch)
}

func (r *GitHubRepositoryInfo) GetBranchSHA(t T, branch string) string {
	var sha string
	For(t).Expect(func(t TT) {
		branchRef, _, err := r.gh.client.Git.GetRef(t, r.Owner, r.Name, "heads/"+branch)
		For(t).Expect(err).Will(BeNil())
		For(t).Expect(branchRef).Will(Not(BeNil()))
		sha = branchRef.GetObject().GetSHA()
		For(t).Expect(sha).Will(Not(BeEmpty()))
	}).Will(Eventually(Succeed()).Within(5 * time.Second).ProbingEvery(time.Second))

	return sha
}

func (r *GitHubRepositoryInfo) CreateFile(t T, branch string) string {
	var sha string
	For(t).Expect(func(t TT) {
		branchRef, _, err := r.gh.client.Repositories.GetBranch(t, r.Owner, r.Name, branch, 0)
		For(t).Expect(err).Will(BeNil())
		For(t).Expect(branchRef).Will(Not(BeNil()))

		file := stringsutil.RandomHash(7) + ".txt"
		cr, _, err := r.gh.client.Repositories.CreateFile(t, r.Owner, r.Name, file, &github.RepositoryContentFileOptions{
			Message: github.String(stringsutil.RandomHash(32)),
			Content: []byte(stringsutil.RandomHash(32)),
			Branch:  &branch,
		})
		For(t).Expect(err).Will(BeNil())

		sha = cr.GetSHA()
		For(t).Expect(sha).Will(Not(BeEmpty()))
	}).Will(Eventually(Succeed()).Within(10 * time.Second).ProbingEvery(time.Second))
	return sha
}

func (r *GitHubRepositoryInfo) DeleteBranch(t T, branch string) {
	For(t).Expect(r.gh.client.Git.DeleteRef(For(t).Context(), r.Owner, r.Name, "heads/"+branch)).Will(Succeed())
}
