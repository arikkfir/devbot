package github_test

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	act "github.com/arikkfir/devbot/backend/internal/controllers/github"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("GitHub Repository Management", Pending, func() {

	When("github repository object is created", func() {
		var gh *github.Client
		BeforeEach(func(ctx context.Context) { CreateGitHubClient(ctx, &gh) })

		var ghRepoName string
		BeforeEach(func(ctx context.Context) { CreateGitHubRepository(ctx, gh, &ghRepoName) })

		var ghRepo *github.Repository
		BeforeEach(func(ctx context.Context) {
			r, _, err := gh.Repositories.Get(ctx, GitHubOwner, ghRepoName)
			Expect(err).ToNot(HaveOccurred())
			ghRepo = r
		})

		var defaultBranchSHA string
		BeforeEach(func(ctx context.Context) {
			GetGitHubBranchCommitSHA(ctx, gh, ghRepo.GetName(), ghRepo.GetDefaultBranch(), &defaultBranchSHA)
		})

		var k client.WithWatch
		BeforeEach(func(ctx context.Context) { CreateKubernetesClient(scheme, &k) })

		var ghAuthSecretNamespace, ghAuthSecretName string
		BeforeEach(func(ctx context.Context) {
			ghAuthSecretNamespace, ghAuthSecretName = "default", strings.RandomHash(7)
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Namespace: ghAuthSecretNamespace, Name: ghAuthSecretName},
				Data:       map[string][]byte{"GITHUB_TOKEN": []byte(os.Getenv("GITHUB_TOKEN"))},
			}
			Expect(k.Create(ctx, secret)).To(Succeed())
			DeferCleanup(func() {
				Expect(k.Delete(ctx, secret)).To(Succeed())
			})
		})

		var name, namespace string
		BeforeEach(func(ctx context.Context) {
			r := &apiv1.GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
				Spec: apiv1.GitHubRepositorySpec{
					Owner: ghRepo.GetOwner().GetName(),
					Name:  ghRepoName,
					Auth: apiv1.GitHubRepositoryAuth{
						PersonalAccessToken: &apiv1.GitHubRepositoryAuthPersonalAccessToken{
							Secret: apiv1.SecretReferenceWithOptionalNamespace{
								Name:      ghAuthSecretName,
								Namespace: ghAuthSecretNamespace,
							},
							Key: "GITHUB_TOKEN",
						},
					},
					RefreshInterval: "15s",
				},
			}
			Expect(k.Create(ctx, r)).To(Succeed())
			DeferCleanup(func() { Expect(k.Delete(context.WithoutCancel(ctx), r)).To(Succeed()) })
		})

		It("should sync github repository object and default branch", func(ctx context.Context) {
			Eventually(func(o Gomega) {
				r := &apiv1.GitHubRepository{}
				o.Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, r)).To(Succeed())
				o.Expect(r.Finalizers).To(ConsistOf(act.Finalizer))
				o.Expect(r.Status.DefaultBranch).To(Equal(ghRepo.GetDefaultBranch()))
				o.Expect(r.Status.GetInvalidCondition()).To(BeNil())
				o.Expect(r.Status.GetUnauthenticatedCondition()).To(BeNil())
				o.Expect(r.Status.GetStaleCondition()).To(BeNil())

				refs := &apiv1.GitHubRepositoryRefList{}
				o.Expect(k.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(scheme, r))).To(Succeed())
				o.Expect(refs.Items).To(HaveLen(1))
				o.Expect(refs.Items[0].Spec.Ref).To(Equal("main"))
				o.Expect(refs.Items[0].Status.RepositoryOwner).To(Equal(r.Spec.Owner))
				o.Expect(refs.Items[0].Status.RepositoryName).To(Equal(r.Spec.Name))
				o.Expect(refs.Items[0].Status.CommitSHA).To(Equal(defaultBranchSHA))
				o.Expect(refs.Items[0].Status.GetInvalidCondition()).To(BeNil())
				o.Expect(refs.Items[0].Status.GetStaleCondition()).To(BeNil())
			}).Within(time.Minute).WithPolling(5 * time.Second).Should(Succeed())
		})

		When("a new branch is created", func() {

			var newBranchName string
			BeforeEach(func(ctx context.Context) {
				newBranchName = strings.RandomHash(7)
				CreateGitHubBranch(ctx, gh, ghRepo.GetName(), newBranchName, true)
			})

			It("should create github repository ref object", func(ctx context.Context) {
				Eventually(func(o Gomega) {
					r := &apiv1.GitHubRepository{}
					o.Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, r)).To(Succeed())

					refs := &apiv1.GitHubRepositoryRefList{}
					o.Expect(k.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(scheme, r))).To(Succeed())
					o.Expect(refs.Items).To(HaveLen(2))
					o.Expect(refs.Items).To(ContainElement(MatchFields(IgnoreExtras, Fields{
						"Spec": MatchFields(IgnoreExtras, Fields{"Ref": Equal(newBranchName)}),
					})))
				}).Within(time.Minute).WithPolling(5 * time.Second).Should(Succeed())
			})

			When("the new branch is updated", func() {
				BeforeEach(func(ctx context.Context) {
					_, _, err := gh.Repositories.CreateFile(ctx, ghRepo.GetOwner().GetName(), ghRepo.GetName(), "README.md", &github.RepositoryContentFileOptions{
						Branch:  github.String(newBranchName),
						Content: []byte(strings.RandomHash(32)),
						Message: github.String(strings.RandomHash(32)),
					})
					Expect(err).ToNot(HaveOccurred())
				})

				It("should sync github repository ref object", func(ctx context.Context) {
					Eventually(func(o Gomega) {
						r := &apiv1.GitHubRepository{}
						o.Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, r)).To(Succeed())

						refs := &apiv1.GitHubRepositoryRefList{}
						o.Expect(k.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(scheme, r))).To(Succeed())

						var ref *apiv1.GitHubRepositoryRef
						for _, r := range refs.Items {
							if r.Spec.Ref == newBranchName {
								ref = &r
								break
							}
						}
						o.Expect(ref).ToNot(BeNil())
						o.Expect(ref.Status.CommitSHA).To(Equal(defaultBranchSHA))
						o.Expect(ref.Status.GetInvalidCondition()).To(BeNil())
						o.Expect(ref.Status.GetStaleCondition()).To(BeNil())
					}).Within(time.Minute).WithPolling(5 * time.Second).Should(Succeed())
				})
			})

			When("branch is subsequently removed", func() {

				BeforeEach(func(ctx context.Context) {
					Expect(gh.Git.DeleteRef(ctx, ghRepo.GetOwner().GetName(), ghRepo.GetName(), "heads/"+newBranchName)).To(Succeed())
				})

				It("should delete github repository ref object", func(ctx context.Context) {
					r := &apiv1.GitHubRepository{}
					Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, r)).To(Succeed())
					Eventually(func(o Gomega) {
						refs := &apiv1.GitHubRepositoryRefList{}
						o.Expect(k.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(scheme, r))).To(Succeed())
						o.Expect(refs.Items).To(HaveLen(1))
						o.Expect(refs.Items[0].Spec.Ref).To(Equal(ghRepo.GetDefaultBranch()))
					}).Within(time.Minute).WithPolling(5 * time.Second).Should(Succeed())
				})
			})

		})
	})
})
