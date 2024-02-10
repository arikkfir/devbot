package e2e_test

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/e2e"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("GitHub Branch Tracking", func() {

	When("github repository object is created", func() {
		var gh *GitHub
		var repo *GitHubRepositoryInfo
		var mainSHA string
		BeforeEach(func(ctx context.Context) {
			gh = NewGitHub(ctx)
			repo = gh.CreateRepository(ctx, "bare")
			mainSHA = repo.GetBranchSHA(ctx, "main")
		})

		var k *Kubernetes
		var ns *Namespace
		var repoObjName string
		var refreshInterval time.Duration
		BeforeEach(func(ctx context.Context) {
			k = NewKubernetes(ctx)
			ns = k.CreateNamespace(ctx)
			ghAuthSecretName, ghAuthSecretKeyName := ns.CreateGitHubAuthSecret(ctx, gh.Token)
			refreshInterval = 5 * time.Second
			ns.CreateGitHubRepository(ctx, &repoObjName, apiv1.GitHubRepositorySpec{
				Owner: repo.Owner,
				Name:  repo.Name,
				Auth: apiv1.GitHubRepositoryAuth{
					PersonalAccessToken: &apiv1.GitHubRepositoryAuthPersonalAccessToken{
						Secret: apiv1.SecretReferenceWithOptionalNamespace{
							Name:      ghAuthSecretName,
							Namespace: ns.Name,
						},
						Key: ghAuthSecretKeyName,
					},
				},
				RefreshInterval: refreshInterval.String(),
			})
		})

		It("should sync github repository object and default branch", func(ctx context.Context) {

			Eventually(func(o Gomega) {
				r := &apiv1.GitHubRepository{}
				o.Expect(k.Client.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: repoObjName}, r)).Error().NotTo(HaveOccurred())
				o.Expect(r.Status.GetFailedToInitializeCondition()).To(BeNil())
				o.Expect(r.Status.GetFinalizingCondition()).To(BeNil())
				o.Expect(r.Status.GetInvalidCondition()).To(BeNil())
				o.Expect(r.Status.GetStaleCondition()).To(BeNil())
				o.Expect(r.Status.GetUnauthenticatedCondition()).To(BeNil())

				refs := &apiv1.GitHubRepositoryRefList{}
				o.Expect(k.Client.List(ctx, refs, client.InNamespace(ns.Name), k8s.OwnedBy(k.Client.Scheme(), r))).Error().NotTo(HaveOccurred())
				o.Expect(refs.Items).To(ConsistOf(BeReady(*r, "main", mainSHA)))
			}).WithContext(ctx).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())

		})

		It("should sync according to refresh interval", func(ctx context.Context) {

			// Wait until initial reconciliation finished
			Eventually(func(o Gomega) {
				r := &apiv1.GitHubRepository{}
				o.Expect(k.Client.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: repoObjName}, r)).Error().NotTo(HaveOccurred())
				o.Expect(r.Status.GetFailedToInitializeCondition()).To(BeNil())
				o.Expect(r.Status.GetFinalizingCondition()).To(BeNil())
				o.Expect(r.Status.GetInvalidCondition()).To(BeNil())
				o.Expect(r.Status.GetStaleCondition()).To(BeNil())
				o.Expect(r.Status.GetUnauthenticatedCondition()).To(BeNil())

				refs := &apiv1.GitHubRepositoryRefList{}
				o.Expect(k.Client.List(ctx, refs, client.InNamespace(ns.Name), k8s.OwnedBy(k.Client.Scheme(), r))).Error().NotTo(HaveOccurred())
				o.Expect(refs.Items).To(ConsistOf(BeReady(*r, "main", mainSHA)))
			}).WithContext(ctx).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())

			// Double the refresh interval
			r := &apiv1.GitHubRepository{}
			Expect(k.Client.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: repoObjName}, r)).Error().NotTo(HaveOccurred())
			r.Spec.RefreshInterval = (refreshInterval * 3).String()
			Expect(k.Client.Update(ctx, r)).Error().NotTo(HaveOccurred())

			// Wait for half of the new refresh interval
			time.Sleep(refreshInterval / 2)

			// Create a new branch
			newBranchName := "b1"
			newBranchSHA := repo.CreateBranch(ctx, "b1")

			// The double refresh interval should not have passed yet - we should be on a single branch still
			rr := &apiv1.GitHubRepository{}
			Expect(k.Client.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: repoObjName}, rr)).Error().NotTo(HaveOccurred())
			refs := &apiv1.GitHubRepositoryRefList{}
			Expect(k.Client.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(k.Client.Scheme(), r))).Error().NotTo(HaveOccurred())
			Expect(refs.Items).To(ConsistOf(BeReady(*r, "main", mainSHA)))

			// Now wait until the refresh interval does pass, and the 2nd branch is synced
			Eventually(func(o Gomega) {
				r := &apiv1.GitHubRepository{}
				o.Expect(k.Client.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: repoObjName}, r)).Error().NotTo(HaveOccurred())

				refs := &apiv1.GitHubRepositoryRefList{}
				o.Expect(k.Client.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(k.Client.Scheme(), r))).Error().NotTo(HaveOccurred())
				o.Expect(refs.Items).To(ConsistOf(
					BeReady(*r, "main", mainSHA),
					BeReady(*r, newBranchName, newBranchSHA),
				))
			}).WithContext(ctx).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())

		})

		When("a new branch is created", func() {

			var newBranchSHA, newBranchName string
			BeforeEach(func(ctx context.Context) {
				newBranchName = "z" + stringsutil.RandomHash(7) // adding "z" prefix to make sure new branch is last on sort
				newBranchSHA = repo.CreateBranch(ctx, newBranchName)
			})

			It("should create github repository ref object", func(ctx context.Context) {
				Eventually(func(o Gomega) {
					r := &apiv1.GitHubRepository{}
					o.Expect(k.Client.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: repoObjName}, r)).Error().NotTo(HaveOccurred())

					refs := &apiv1.GitHubRepositoryRefList{}
					o.Expect(k.Client.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(k.Client.Scheme(), r))).Error().NotTo(HaveOccurred())
					o.Expect(refs.Items).To(ConsistOf(
						BeReady(*r, "main", mainSHA),
						BeReady(*r, newBranchName, newBranchSHA),
					))
				}).WithContext(ctx).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
			})

			When("the new branch is updated", func() {

				var updatedBranchSHA string
				BeforeEach(func(ctx context.Context) { updatedBranchSHA = repo.CreateFile(ctx, newBranchName) })

				It("should sync github repository ref object", func(ctx context.Context) {
					Eventually(func(o Gomega) {
						r := &apiv1.GitHubRepository{}
						o.Expect(k.Client.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: repoObjName}, r)).Error().NotTo(HaveOccurred())

						refs := &apiv1.GitHubRepositoryRefList{}
						o.Expect(k.Client.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(k.Client.Scheme(), r))).Error().NotTo(HaveOccurred())
						o.Expect(refs.Items).To(ConsistOf(
							BeReady(*r, "main", mainSHA),
							BeReady(*r, newBranchName, updatedBranchSHA),
						))
					}).WithContext(ctx).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
				})
			})

			When("branch is subsequently removed", func() {

				BeforeEach(func(ctx context.Context) { repo.DeleteBranch(ctx, newBranchName) })

				It("should delete github repository ref object", func(ctx context.Context) {
					Eventually(func(o Gomega) {
						r := &apiv1.GitHubRepository{}
						o.Expect(k.Client.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: repoObjName}, r)).Error().NotTo(HaveOccurred())

						refs := &apiv1.GitHubRepositoryRefList{}
						o.Expect(k.Client.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(k.Client.Scheme(), r))).Error().NotTo(HaveOccurred())
						o.Expect(refs.Items).To(ConsistOf(BeReady(*r, "main", mainSHA)))
					}).WithContext(ctx).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
				})
			})

		})
	})
})
