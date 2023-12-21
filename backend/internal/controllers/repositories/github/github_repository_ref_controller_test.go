package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	k8sutil "github.com/arikkfir/devbot/backend/internal/util/k8s"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe(apiv1.GitHubRepositoryRefGVK.Kind, func() {
	var k8s client.Client
	var gh *github.Client
	var repoName, mainSHA string

	BeforeEach(func(ctx context.Context) { CreateKubernetesClient(&k8s) })
	BeforeEach(func(ctx context.Context) { CreateGitHubClient(ctx, k8s, &gh) })
	BeforeEach(func(ctx context.Context) { CreateGitHubRepository(ctx, k8s, gh, &repoName) })
	BeforeEach(func(ctx context.Context) { GetGitHubBranchCommitSHA(ctx, gh, repoName, "main", &mainSHA) })

	Describe("initialization", func() {
		var crNamespace, crName string
		BeforeEach(func(ctx context.Context) { CreateGitHubRepositoryCR(ctx, k8s, repoName, &crName, &crNamespace) })
		It("should initialize the ref object", func(ctx context.Context) {
			Eventually(func(o Gomega) {
				r, refs := &apiv1.GitHubRepository{}, &apiv1.GitHubRepositoryRefList{}
				o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
				o.Expect(k8sutil.GetOwnedChildrenManually(ctx, k8s, r, refs)).To(Succeed())
				o.Expect(refs.Items).To(HaveLen(1))
				o.Expect(refs.Items[0].Finalizers).To(ContainElement(RepositoryRefFinalizer))
				o.Expect(refs.Items[0].Spec.Ref).To(Equal("refs/heads/main"))
				o.Expect(refs.Items[0].Status.CommitSHA).To(Equal(mainSHA))
				o.Expect(refs.Items[0].Status.Conditions).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(apiv1.ConditionTypeCurrent),
					"Status":  Equal(metav1.ConditionTrue),
					"Reason":  Equal(apiv1.ReasonSynced),
					"Message": Equal("Commit SHA up-to-date"),
				}), MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(apiv1.ConditionTypeAuthenticatedToGitHub),
					"Status":  Equal(metav1.ConditionTrue),
					"Reason":  Equal(apiv1.ReasonAuthenticated),
					"Message": Equal("Authenticated to GitHub"),
				})))
			}, 30*time.Second, 1*time.Second).Should(Succeed())
		})
	})
	Describe("branch update", func() {
		When("when branch is updated", func() {
			var crNamespace, crName string
			BeforeEach(func(ctx context.Context) { CreateGitHubRepositoryCR(ctx, k8s, repoName, &crName, &crNamespace) })
			It("should detect the ref SHA change", func(ctx context.Context) {
				Eventually(func(o Gomega) {
					r, refs := &apiv1.GitHubRepository{}, &apiv1.GitHubRepositoryRefList{}
					o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
					o.Expect(k8sutil.GetOwnedChildrenManually(ctx, k8s, r, refs)).To(Succeed())
					o.Expect(refs.Items).To(HaveLen(1))
					o.Expect(refs.Items[0].Spec.Ref).To(Equal("refs/heads/main"))
					o.Expect(refs.Items[0].Status.CommitSHA).To(Equal(mainSHA))
					o.Expect(refs.Items[0].Status.Conditions).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
						"Type":    Equal(apiv1.ConditionTypeCurrent),
						"Status":  Equal(metav1.ConditionTrue),
						"Reason":  Equal(apiv1.ReasonSynced),
						"Message": Equal("Commit SHA up-to-date"),
					}), MatchFields(IgnoreExtras, Fields{
						"Type":    Equal(apiv1.ConditionTypeAuthenticatedToGitHub),
						"Status":  Equal(metav1.ConditionTrue),
						"Reason":  Equal(apiv1.ReasonAuthenticated),
						"Message": Equal("Authenticated to GitHub"),
					})))
				})

				var newSHA string
				CreateGitHubFile(ctx, gh, repoName, "main", &newSHA)

				Eventually(func(o Gomega) {
					r, refs := &apiv1.GitHubRepository{}, &apiv1.GitHubRepositoryRefList{}
					o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
					o.Expect(k8sutil.GetOwnedChildrenManually(ctx, k8s, r, refs)).To(Succeed())
					o.Expect(refs.Items[0].Spec.Ref).To(Equal("refs/heads/main"))
					o.Expect(refs.Items[0].Status.CommitSHA).To(Equal(newSHA))
					o.Expect(refs.Items[0].Status.Conditions).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
						"Type":    Equal(apiv1.ConditionTypeCurrent),
						"Status":  Equal(metav1.ConditionTrue),
						"Reason":  Equal(apiv1.ReasonSynced),
						"Message": Equal("Commit SHA up-to-date"),
					}), MatchFields(IgnoreExtras, Fields{
						"Type":    Equal(apiv1.ConditionTypeAuthenticatedToGitHub),
						"Status":  Equal(metav1.ConditionTrue),
						"Reason":  Equal(apiv1.ReasonAuthenticated),
						"Message": Equal("Authenticated to GitHub"),
					})))
				})
			})
		})
	})
	Describe("branch deletion", func() {
		When("when branch is deleted", func() {
			var crNamespace, crName string
			BeforeEach(func(ctx context.Context) { CreateGitHubRepositoryCR(ctx, k8s, repoName, &crName, &crNamespace) })
			It("should delete the ref object", func(ctx context.Context) {

				Eventually(func(o Gomega) {
					r, refs := &apiv1.GitHubRepository{}, &apiv1.GitHubRepositoryRefList{}
					o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
					o.Expect(k8sutil.GetOwnedChildrenManually(ctx, k8s, r, refs)).To(Succeed())
					o.Expect(refs.Items).To(HaveLen(1))
					o.Expect(refs.Items[0].Spec.Ref).To(Equal("refs/heads/main"))
					o.Expect(refs.Items[0].Status.CommitSHA).To(Equal(mainSHA))
				}, 2*time.Minute, 1*time.Second).Should(Succeed())

				DeleteGitHubBranch(ctx, gh, repoName, "main")

				Eventually(func(o Gomega) {
					r, refs := &apiv1.GitHubRepository{}, &apiv1.GitHubRepositoryRefList{}
					o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
					o.Expect(k8sutil.GetOwnedChildrenManually(ctx, k8s, r, refs)).To(Succeed())
					o.Expect(refs.Items).To(BeEmpty())
				}, 2*time.Minute, 1*time.Second).Should(Succeed())
			})
		})
	})
})
