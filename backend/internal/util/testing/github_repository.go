package testing

import (
	"bytes"
	"context"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"github.com/google/go-github/v56/github"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
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

func (r *GitHubRepositoryInfo) SetupWebhook(ctx context.Context, t T) {
	if r.WebhookSecret != "" {
		return
	}

	webhookSecret := stringsutil.RandomHash(16)

	var smeeCommand *exec.Cmd
	smeeOutput := bytes.Buffer{}
	smeeCommand = exec.Command("smee", "--port", "8080", "--path", "/github/webhook")
	smeeCommand.Stdout = &smeeOutput
	smeeCommand.Stderr = os.Stderr
	With(t).Verify(smeeCommand.Start()).Will(Succeed()) // TODO: not evaluated, ensure this fails tests
	t.Cleanup(func() { With(t).Verify(smeeCommand.Process.Signal(os.Interrupt)).Will(Succeed()) })

	With(t).Verify(smeeCommand.Stdout).Will(Say(smeeTunnelRE)).Within(10*time.Second, 100*time.Millisecond)
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
	With(t).Verify(err).Will(BeNil()) // TODO: not evaluated, ensure this fails tests
	t.Cleanup(func() {
		With(t).Verify(r.gh.client.Repositories.DeleteHook(ctx, r.Owner, r.Name, hook.GetID())).Will(Succeed()) // TODO: not evaluated, ensure this fails tests
	})

	r.WebhookSecret = webhookSecret
}

func (r *GitHubRepositoryInfo) CreateBranch(ctx context.Context, t T, branch string) string {
	mainRef, _, err := r.gh.client.Git.GetRef(ctx, r.Owner, r.Name, "heads/main")
	With(t).Verify(err).Will(BeNil()).OrFail()
	With(t).Verify(mainRef).Will(Not(BeNil())).OrFail()
	With(t).Verify(r.gh.client.Git.CreateRef(ctx, r.Owner, r.Name, &github.Reference{
		Ref:    github.String("refs/heads/" + branch),
		Object: mainRef.Object,
	})).Will(Succeed()).OrFail()
	return r.GetBranchSHA(ctx, t, branch)
}

func (r *GitHubRepositoryInfo) GetBranchSHA(ctx context.Context, t T, branch string) string {
	var sha string
	With(t).Verify(func(t T) {
		branchRef, _, err := r.gh.client.Git.GetRef(ctx, r.Owner, r.Name, "heads/"+branch)
		With(t).Verify(err).Will(BeNil()).OrFail()
		With(t).Verify(branchRef).Will(Not(BeNil())).OrFail()
		sha = branchRef.GetObject().GetSHA()
		With(t).Verify(sha).Will(Not(BeEmpty())).OrFail()
	}).Will(Succeed()).Within(5*time.Second, 1*time.Second)

	return sha
}

func (r *GitHubRepositoryInfo) CreateFile(ctx context.Context, t T, branch string) string {
	var sha string
	With(t).Verify(func(t T) {
		branchRef, _, err := r.gh.client.Repositories.GetBranch(ctx, r.Owner, r.Name, branch, 0)
		With(t).Verify(err).Will(BeNil()).OrFail()
		With(t).Verify(branchRef).Will(Not(BeNil())).OrFail()

		file := stringsutil.RandomHash(7) + ".txt"
		cr, _, err := r.gh.client.Repositories.CreateFile(ctx, r.Owner, r.Name, file, &github.RepositoryContentFileOptions{
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

func (r *GitHubRepositoryInfo) DeleteBranch(ctx context.Context, t T, branch string) {
	With(t).Verify(r.gh.client.Git.DeleteRef(ctx, r.Owner, r.Name, "heads/"+branch)).Will(Succeed())
}
