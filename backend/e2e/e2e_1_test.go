package e2e_test

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/e2e"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("GitHub Branch Tracking", func() {

	When("github repository object is created", func() {
		var gh *github.Client
		var ghAuthToken, ghOwner, ghRepoName, mainSHA string
		BeforeEach(func(ctx context.Context) {
			CreateGitHubClient(ctx, &gh, &ghAuthToken)
			CreateGitHubRepository(ctx, gh, "bare", &ghOwner, &ghRepoName, &mainSHA)
		})

		var k client.Client
		BeforeEach(func(ctx context.Context) { CreateKubernetesClient(ctx, scheme, &k) })

		var namespace string
		BeforeEach(func(ctx context.Context) { CreateKubernetesNamespace(ctx, k, &namespace) })

		var ghAuthSecretName, ghAuthSecretKeyName string
		BeforeEach(func(ctx context.Context) {
			CreateGitHubSecretForDevbot(ctx, k, namespace, ghAuthToken, &ghAuthSecretName, &ghAuthSecretKeyName)
		})

		var repoObjName string
		BeforeEach(func(ctx context.Context) { repoObjName = stringsutil.RandomHash(7) })

		BeforeEach(func(ctx context.Context) {
			r := &apiv1.GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: repoObjName},
				Spec: apiv1.GitHubRepositorySpec{
					Owner: ghOwner,
					Name:  ghRepoName,
					Auth: apiv1.GitHubRepositoryAuth{
						PersonalAccessToken: &apiv1.GitHubRepositoryAuthPersonalAccessToken{
							Secret: apiv1.SecretReferenceWithOptionalNamespace{
								Name:      ghAuthSecretName,
								Namespace: namespace,
							},
							Key: ghAuthSecretKeyName,
						},
					},
					RefreshInterval: "60s",
				},
			}
			Expect(k.Create(ctx, r)).To(Succeed())
			DeferCleanup(func(ctx context.Context) { Expect(k.Delete(ctx, r)).To(Succeed()) })
		})

		It("should sync github repository object and default branch", func(ctx context.Context) {
			Eventually(func(o Gomega) {
				r := &apiv1.GitHubRepository{}
				o.Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: repoObjName}, r)).To(Succeed())
				o.Expect(r.Status.GetFailedToInitializeCondition()).To(BeNil())
				o.Expect(r.Status.GetFinalizingCondition()).To(BeNil())
				o.Expect(r.Status.GetInvalidCondition()).To(BeNil())
				o.Expect(r.Status.GetStaleCondition()).To(BeNil())
				o.Expect(r.Status.GetUnauthenticatedCondition()).To(BeNil())

				refs := &apiv1.GitHubRepositoryRefList{}
				o.Expect(k.List(ctx, refs, client.InNamespace(namespace), k8s.OwnedBy(scheme, r))).To(Succeed())
				o.Expect(refs.Items).To(ConsistOf(BeReady(*r, "main", mainSHA)))
			}).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
		})

		When("a new branch is created", func() {

			var newBranchSHA, newBranchName string
			BeforeEach(func(ctx context.Context) {
				newBranchName = "z" + stringsutil.RandomHash(7) // adding "z" prefix to make sure new branch is last on sort
				CreateGitHubBranch(ctx, gh, ghOwner, ghRepoName, newBranchName, &newBranchSHA)
			})

			It("should create github repository ref object", func(ctx context.Context) {
				Eventually(func(o Gomega) {
					r := &apiv1.GitHubRepository{}
					o.Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: repoObjName}, r)).To(Succeed())

					refs := &apiv1.GitHubRepositoryRefList{}
					o.Expect(k.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(scheme, r))).To(Succeed())
					o.Expect(refs.Items).To(ConsistOf(
						BeReady(*r, "main", mainSHA),
						BeReady(*r, newBranchName, newBranchSHA),
					))
				}).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
			})

			When("the new branch is updated", func() {
				var updatedBranchSHA string
				BeforeEach(func(ctx context.Context) {
					CreateGitHubFile(ctx, gh, ghOwner, ghRepoName, newBranchName, &updatedBranchSHA)
				})

				It("should sync github repository ref object", func(ctx context.Context) {
					Eventually(func(o Gomega) {
						r := &apiv1.GitHubRepository{}
						o.Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: repoObjName}, r)).To(Succeed())

						refs := &apiv1.GitHubRepositoryRefList{}
						o.Expect(k.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(scheme, r))).To(Succeed())
						o.Expect(refs.Items).To(ConsistOf(
							BeReady(*r, "main", mainSHA),
							BeReady(*r, newBranchName, updatedBranchSHA),
						))
					}).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
				})
			})

			When("branch is subsequently removed", func() {

				BeforeEach(func(ctx context.Context) {
					DeleteGitHubBranch(ctx, gh, ghOwner, ghRepoName, newBranchName)
				})

				It("should delete github repository ref object", func(ctx context.Context) {
					Eventually(func(o Gomega) {
						r := &apiv1.GitHubRepository{}
						o.Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: repoObjName}, r)).To(Succeed())

						refs := &apiv1.GitHubRepositoryRefList{}
						o.Expect(k.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(scheme, r))).To(Succeed())
						o.Expect(refs.Items).To(ConsistOf(BeReady(*r, "main", mainSHA)))
					}).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
				})
			})
		})
	})
})
