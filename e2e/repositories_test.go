package e2e_test

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/arikkfir/devbot/api/v1"
	. "github.com/arikkfir/devbot/e2e/matchers"
	"github.com/arikkfir/devbot/e2e/util"
	"github.com/arikkfir/devbot/internal/util/strings"
)

var _ = Describe("Repository Reconciliation", func() {
	var gh *github.Client
	BeforeEach(func(ctx context.Context) { gh = util.NewGitHubClient(ctx) })

	var c client.Client
	var rc *rest.Config
	BeforeEach(func() { c, rc = util.NewK8sClient() })

	var nsName string
	BeforeEach(func(ctx context.Context) { nsName = util.CreateK8sNamespace(ctx, c) })
	JustAfterEach(func(ctx SpecContext) { util.PrintK8sDebugInfo(ctx, c, rc, nsName) })

	var ghRepo *github.Repository
	BeforeEach(func(ctx context.Context) {
		ghRepo = util.CreateGitHubRepository(ctx, gh, repositoriesFS, "repositories/bare")
	})

	var tokenSecretName, tokenSecretKey, token string
	BeforeEach(func(ctx context.Context) { token = util.GetGitHubToken() })
	BeforeEach(func(ctx context.Context) {
		tokenSecretName, tokenSecretKey = util.CreateK8sSecretWithGitHubAuthToken(ctx, c, nsName, token)
	})
	BeforeEach(func(ctx context.Context) { util.GrantK8sAccessToSecret(ctx, c, nsName, tokenSecretName) }) // Need permissions to List secrets

	var kRepoName string
	BeforeEach(func(ctx context.Context) {
		kRepoName = util.CreateK8sRepository(ctx, c, nsName, apiv1.RepositorySpec{
			GitHub: &apiv1.GitHubRepositorySpec{
				Owner: ghRepo.Owner.GetLogin(), Name: ghRepo.GetName(),
				PersonalAccessToken: apiv1.GitHubRepositoryPersonalAccessToken{
					Secret: apiv1.SecretReferenceWithOptionalNamespace{Name: tokenSecretName, Namespace: nsName},
					Key:    tokenSecretKey,
				},
			},
			RefreshInterval: "5s",
		})
	})

	It("should set resolved repository name in Repository status", func(ctx context.Context) {
		Eventually(func(g Gomega) {
			repo := &apiv1.Repository{}
			g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
			g.Expect(repo.Status.ResolvedName).To(Equal(repo.Spec.GitHub.Owner + "/" + repo.Spec.GitHub.Name))
		}, "30s").Should(Succeed())
	})

	Describe("refresh interval", func() {
		When("refresh interval is empty", func() {
			BeforeEach(func(ctx context.Context) {
				repo := &apiv1.Repository{ObjectMeta: v1.ObjectMeta{Namespace: nsName, Name: kRepoName}}
				util.PatchK8sObject(ctx, c, repo, util.JSONPatchItem{Op: util.JSONPatchOperationRemove, Path: "/spec/refreshInterval"})
			})
			It("should use the default refresh interval value", func(ctx context.Context) {
				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					g.Expect(repo.Spec.RefreshInterval).To(Equal("5m"))
					g.Expect(repo.Status.Conditions).To(BeEmpty())
				}, "1m").Should(Succeed())
			})
		})
		When("refresh interval is invalid", func() {
			BeforeEach(func(ctx context.Context) {
				repo := &apiv1.Repository{ObjectMeta: v1.ObjectMeta{Namespace: nsName, Name: kRepoName}}
				util.PatchK8sObject(ctx, c, repo, util.JSONPatchItem{Op: util.JSONPatchOperationReplace, Path: "/spec/refreshInterval", Value: "abc"})
			})
			It("should be marked as invalid and stale", func(ctx context.Context) {
				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					const invalidDurationMessage = `time: invalid duration "abc"`
					g.Expect(repo.Status.Conditions).To(ConsistOf(
						ConditionWith(Type(apiv1.Invalid), Status(v1.ConditionTrue), Reason(apiv1.InvalidRefreshInterval), Message(Equal(invalidDurationMessage))),
						ConditionWith(Type(apiv1.Stale), Status(v1.ConditionUnknown), Reason(apiv1.Invalid), Message(Equal(invalidDurationMessage))),
					), "conditions state are incorrect")
				}, "30s").Should(Succeed())
			})
		})
		When("refresh interval is too low", func() {
			BeforeEach(func(ctx context.Context) {
				repo := &apiv1.Repository{ObjectMeta: v1.ObjectMeta{Namespace: nsName, Name: kRepoName}}
				util.PatchK8sObject(ctx, c, repo, util.JSONPatchItem{Op: util.JSONPatchOperationReplace, Path: "/spec/refreshInterval", Value: "2s"})
			})
			It("should be marked as invalid and stale", func(ctx context.Context) {
				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					const invalidDurationMessage = `refresh interval '2s' is too low (must not be less than 5s)`
					g.Expect(repo.Status.Conditions).To(ConsistOf(
						ConditionWith(Type(apiv1.Invalid), Status(v1.ConditionTrue), Reason(apiv1.InvalidRefreshInterval), Message(Equal(invalidDurationMessage))),
						ConditionWith(Type(apiv1.Stale), Status(v1.ConditionUnknown), Reason(apiv1.Invalid), Message(Equal(invalidDurationMessage))),
					), "conditions state are incorrect")
				}, "30s").Should(Succeed())
			})
		})
		When("refresh interval is valid", func() {
			BeforeEach(func(ctx context.Context) {
				repo := &apiv1.Repository{ObjectMeta: v1.ObjectMeta{Namespace: nsName, Name: kRepoName}}
				util.PatchK8sObject(ctx, c, repo, util.JSONPatchItem{Op: util.JSONPatchOperationReplace, Path: "/spec/refreshInterval", Value: "6s"})
			})
			It("should have no conditions", func(ctx context.Context) {
				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					g.Expect(repo.Status.Conditions).To(BeEmpty())
				}, "30s").Should(Succeed())
			})
		})
	})
	Describe("personal access token", func() {
		When("the k8s secret cannot be found", func() {
			var bogusSecretName string
			BeforeEach(func(ctx context.Context) {
				bogusSecretName = strings.RandomHash(7)
				repo := &apiv1.Repository{ObjectMeta: v1.ObjectMeta{Namespace: nsName, Name: kRepoName}}
				util.PatchK8sObject(ctx, c, repo, util.JSONPatchItem{Op: util.JSONPatchOperationReplace, Path: "/spec/github/personalAccessToken/secret/name", Value: bogusSecretName})
				util.GrantK8sAccessToSecret(ctx, c, nsName, repo.Spec.GitHub.PersonalAccessToken.Secret.Name)
			})
			It("should be marked as unauthenticated and stale", func(ctx context.Context) {
				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					invalidDurationMessage := fmt.Sprintf("Secret '%s/%s' not found", nsName, bogusSecretName)
					g.Expect(repo.Status.Conditions).To(ConsistOf(
						ConditionWith(Type(apiv1.Unauthenticated), Status(v1.ConditionTrue), Reason(apiv1.AuthSecretNotFound), Message(Equal(invalidDurationMessage))),
						ConditionWith(Type(apiv1.Stale), Status(v1.ConditionUnknown), Reason(apiv1.Unauthenticated), Message(Equal(invalidDurationMessage))),
					), "conditions state are incorrect")
				}, "30s").Should(Succeed())
			})
		})
		When("the k8s secret is not accessible", func() {
			var newTokenSecretName, newTokenSecretKey string
			BeforeEach(func(ctx context.Context) {
				newTokenSecretName, newTokenSecretKey = util.CreateK8sSecretWithGitHubAuthToken(ctx, c, nsName, token)
				repo := &apiv1.Repository{ObjectMeta: v1.ObjectMeta{Namespace: nsName, Name: kRepoName}}
				util.PatchK8sObject(ctx, c, repo,
					util.JSONPatchItem{Op: util.JSONPatchOperationReplace, Path: "/spec/github/personalAccessToken/secret/name", Value: newTokenSecretName},
					util.JSONPatchItem{Op: util.JSONPatchOperationReplace, Path: "/spec/github/personalAccessToken/key", Value: newTokenSecretKey},
				)
			})
			It("should be marked as unauthenticated and stale", func(ctx context.Context) {
				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					invalidDurationMessage := fmt.Sprintf("Secret '%s/%s' is not accessible: secrets \"%[2]s\" is forbidden: .*", nsName, newTokenSecretName)
					g.Expect(repo.Status.Conditions).To(ConsistOf(
						ConditionWith(Type(apiv1.Unauthenticated), Status(v1.ConditionTrue), Reason(apiv1.AuthSecretForbidden), Message(MatchRegexp(invalidDurationMessage))),
						ConditionWith(Type(apiv1.Stale), Status(v1.ConditionUnknown), Reason(apiv1.Unauthenticated), Message(MatchRegexp(invalidDurationMessage))),
					), "conditions state are incorrect")
				}, "30s").Should(Succeed())
			})
		})
		When("the k8s secret does not have the specified key", func() {
			var bogusTokenSecretKey string
			BeforeEach(func(ctx context.Context) {
				bogusTokenSecretKey = strings.RandomHash(7)
				repo := &apiv1.Repository{ObjectMeta: v1.ObjectMeta{Namespace: nsName, Name: kRepoName}}
				util.PatchK8sObject(ctx, c, repo,
					util.JSONPatchItem{Op: util.JSONPatchOperationReplace, Path: "/spec/github/personalAccessToken/key", Value: bogusTokenSecretKey},
				)
			})
			It("should be marked as unauthenticated and stale", func(ctx context.Context) {
				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					invalidDurationMessage := fmt.Sprintf("Key '%s' not found in secret '%s/%s'", bogusTokenSecretKey, nsName, tokenSecretName)
					g.Expect(repo.Status.Conditions).To(ConsistOf(
						ConditionWith(Type(apiv1.Unauthenticated), Status(v1.ConditionTrue), Reason(apiv1.AuthSecretKeyNotFound), Message(MatchRegexp(invalidDurationMessage))),
						ConditionWith(Type(apiv1.Stale), Status(v1.ConditionUnknown), Reason(apiv1.Unauthenticated), Message(MatchRegexp(invalidDurationMessage))),
					), "conditions state are incorrect")
				}, "30s").Should(Succeed())
			})
		})
		When("the token in the k8s secret is empty", func() {
			BeforeEach(func(ctx context.Context) {
				encodedValue := base64.StdEncoding.EncodeToString([]byte(""))
				repo := &corev1.Secret{ObjectMeta: v1.ObjectMeta{Namespace: nsName, Name: tokenSecretName}}
				util.PatchK8sObject(ctx, c, repo,
					util.JSONPatchItem{Op: util.JSONPatchOperationReplace, Path: "/data/" + tokenSecretKey, Value: encodedValue},
				)
			})
			It("should be marked as unauthenticated and stale", func(ctx context.Context) {
				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					invalidDurationMessage := fmt.Sprintf("Token in key '%s' in secret '%s/%s' is empty", tokenSecretKey, nsName, tokenSecretName)
					g.Expect(repo.Status.Conditions).To(ConsistOf(
						ConditionWith(Type(apiv1.Unauthenticated), Status(v1.ConditionTrue), Reason(apiv1.AuthTokenEmpty), Message(MatchRegexp(invalidDurationMessage))),
						ConditionWith(Type(apiv1.Stale), Status(v1.ConditionUnknown), Reason(apiv1.Unauthenticated), Message(MatchRegexp(invalidDurationMessage))),
					), "conditions state are incorrect")
				}, "30s").Should(Succeed())
			})
		})
		When("the token in the k8s secret is invalid", func() {
			BeforeEach(func(ctx context.Context) {
				encodedValue := base64.StdEncoding.EncodeToString([]byte(strings.RandomHash(7)))
				repo := &corev1.Secret{ObjectMeta: v1.ObjectMeta{Namespace: nsName, Name: tokenSecretName}}
				util.PatchK8sObject(ctx, c, repo,
					util.JSONPatchItem{Op: util.JSONPatchOperationReplace, Path: "/data/" + tokenSecretKey, Value: encodedValue},
				)
			})
			It("should be marked as unauthenticated and stale", func(ctx context.Context) {
				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					invalidDurationMessage := "Validation request failed: .+"
					g.Expect(repo.Status.Conditions).To(ConsistOf(
						ConditionWith(Type(apiv1.Unauthenticated), Status(v1.ConditionTrue), Reason(apiv1.AuthenticationFailed), Message(MatchRegexp(invalidDurationMessage))),
						ConditionWith(Type(apiv1.Stale), Status(v1.ConditionUnknown), Reason(apiv1.Unauthenticated), Message(MatchRegexp(invalidDurationMessage))),
					), "conditions state are incorrect")
				}, "30s").Should(Succeed())
			})
		})
		When("the token in the k8s secret is valid", func() {
			It("reconciliation succeeds", func(ctx context.Context) {
				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					g.Expect(repo.Status.Conditions).To(BeEmpty())
				}, "30s").Should(Succeed())
			})
		})
	})

	Describe("branch and revisions reconciliation", func() {
		var defaultBranch, defaultBranchSHA string
		BeforeEach(func(ctx context.Context) {
			defaultBranch = ghRepo.GetDefaultBranch()
			defaultBranchSHA = util.GetGitHubRepositoryBranchSHA(ctx, gh, ghRepo, defaultBranch)
		})

		It("should resolve default branch and revision", func(ctx context.Context) {
			Eventually(func(g Gomega) {
				repo := &apiv1.Repository{}
				g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
				g.Expect(repo.Status.Conditions).To(BeEmpty())
				g.Expect(repo.Status.DefaultBranch).To(Equal(defaultBranch))
				g.Expect(repo.Status.Revisions).To(Equal(map[string]string{defaultBranch: defaultBranchSHA}))
			}, "30s").Should(Succeed())
		})

		When("the default branch is updated", func() {
			It("should resolve updated revision", func(ctx context.Context) {
				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					g.Expect(repo.Status.Conditions).To(BeEmpty())
					g.Expect(repo.Status.DefaultBranch).To(Equal(defaultBranch))
					g.Expect(repo.Status.Revisions).To(Equal(map[string]string{defaultBranch: defaultBranchSHA}))
				}, "30s").Should(Succeed())

				updatedDefaultBranchSHA := util.CreateFileInGitHubRepositoryBranch(ctx, gh, ghRepo, defaultBranch)
				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					g.Expect(repo.Status.Conditions).To(BeEmpty())
					g.Expect(repo.Status.DefaultBranch).To(Equal(defaultBranch))
					g.Expect(repo.Status.Revisions).To(Equal(map[string]string{defaultBranch: updatedDefaultBranchSHA}))
				}, "30s").Should(Succeed())
			})
		})

		When("a new branch is created", func() {
			It("should detect and resolve new branch", func(ctx context.Context) {
				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					g.Expect(repo.Status.Conditions).To(BeEmpty())
					g.Expect(repo.Status.DefaultBranch).To(Equal(defaultBranch))
					g.Expect(repo.Status.Revisions).To(Equal(map[string]string{defaultBranch: defaultBranchSHA}))
				}, "30s").Should(Succeed())

				var newBranchName = strings.RandomHash(7)
				var newBranchSHA string
				util.CreateGitHubRepositoryBranch(ctx, gh, ghRepo, newBranchName)
				newBranchSHA = util.GetGitHubRepositoryBranchSHA(ctx, gh, ghRepo, newBranchName)

				Eventually(func(g Gomega) {
					repo := &apiv1.Repository{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kRepoName}, repo)).To(Succeed())
					g.Expect(repo.Status.Conditions).To(BeEmpty())
					g.Expect(repo.Status.DefaultBranch).To(Equal(defaultBranch))
					g.Expect(repo.Status.Revisions).To(Equal(map[string]string{
						defaultBranch: defaultBranchSHA,
						newBranchName: newBranchSHA,
					}), "revisions are wrong")
				}, "30s").Should(Succeed())
			})
		})
	})
})
