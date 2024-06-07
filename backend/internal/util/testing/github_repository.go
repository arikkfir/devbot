package testing

import (
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	. "github.com/arikkfir/justest"
	"github.com/google/go-github/v56/github"

	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
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

func (r *GitHubRepositoryInfo) Delete(t T) {
	With(t).Verify(r.gh.client.Repositories.Delete(r.gh.ctx, r.Owner, r.Name)).Will(Succeed()).OrFail()
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
	With(t).Verify(smeeCommand.Start()).Will(Succeed()).OrFail()
	t.Cleanup(func() { With(t).Verify(smeeCommand.Process.Signal(os.Interrupt)).Will(Succeed()).OrFail() })

	With(t).Verify(smeeCommand.Stdout).Will(Say(smeeTunnelRE)).Within(10*time.Second, 100*time.Millisecond)
	webHookURL := strings.TrimSpace(smeeTunnelRE.FindStringSubmatch(smeeOutput.String())[1])

	hook, _, err := r.gh.client.Repositories.CreateHook(r.gh.ctx, r.Owner, r.Name, &github.Hook{
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
	With(t).Verify(err).Will(BeNil()).OrFail()
	t.Cleanup(func() {
		With(t).Verify(r.gh.client.Repositories.DeleteHook(r.gh.ctx, r.Owner, r.Name, hook.GetID())).Will(Succeed()).OrFail()
	})

	r.WebhookSecret = webhookSecret
}

func (r *GitHubRepositoryInfo) CreateBranch(t T, branch string) string {
	mainRef, _, err := r.gh.client.Git.GetRef(r.gh.ctx, r.Owner, r.Name, "heads/main")
	With(t).Verify(err).Will(BeNil()).OrFail()
	With(t).Verify(mainRef).Will(Not(BeNil())).OrFail()
	With(t).Verify(r.gh.client.Git.CreateRef(r.gh.ctx, r.Owner, r.Name, &github.Reference{
		Ref:    github.String("refs/heads/" + branch),
		Object: mainRef.Object,
	})).Will(Succeed()).OrFail()
	return r.GetBranchSHA(t, branch)
}

func (r *GitHubRepositoryInfo) GetBranchSHA(t T, branch string) string {
	var sha string
	With(t).Verify(func(t T) {
		branchRef, _, err := r.gh.client.Git.GetRef(r.gh.ctx, r.Owner, r.Name, "heads/"+branch)
		With(t).Verify(err).Will(BeNil()).OrFail()
		With(t).Verify(branchRef).Will(Not(BeNil())).OrFail()
		sha = branchRef.GetObject().GetSHA()
		With(t).Verify(sha).Will(Not(BeEmpty())).OrFail()
	}).Will(Succeed()).Within(5*time.Second, 1*time.Second)

	return sha
}

func (r *GitHubRepositoryInfo) CreateFile(t T, branch string) string {
	var sha string
	With(t).Verify(func(t T) {
		branchRef, _, err := r.gh.client.Repositories.GetBranch(r.gh.ctx, r.Owner, r.Name, branch, 0)
		With(t).Verify(err).Will(BeNil()).OrFail()
		With(t).Verify(branchRef).Will(Not(BeNil())).OrFail()

		file := stringsutil.RandomHash(7) + ".txt"
		cr, _, err := r.gh.client.Repositories.CreateFile(r.gh.ctx, r.Owner, r.Name, file, &github.RepositoryContentFileOptions{
			Message: github.String(stringsutil.RandomHash(32)),
			Content: []byte(stringsutil.RandomHash(32)),
			Branch:  &branch,
		})
		With(t).Verify(err).Will(BeNil()).OrFail()

		sha = cr.GetSHA()
		With(t).Verify(sha).Will(Not(BeEmpty())).OrFail()
	}).Will(Succeed()).Within(10*time.Second, 1*time.Second)
	return sha
}

func (r *GitHubRepositoryInfo) DeleteBranch(t T, branch string) {
	With(t).Verify(r.gh.client.Git.DeleteRef(r.gh.ctx, r.Owner, r.Name, "heads/"+branch)).Will(Succeed()).OrFail()
}
