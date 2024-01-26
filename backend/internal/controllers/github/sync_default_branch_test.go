package github_test

import (
	"context"
	. "github.com/arikkfir/devbot/backend/api/v1"
	act "github.com/arikkfir/devbot/backend/internal/controllers/github"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("SyncDefaultBranch", func() {
	var namespace, repoObjName, repoName string
	var k client.Client

	BeforeEach(func(ctx context.Context) {
		namespace = "default"
		repoObjName = strings.RandomHash(7)
		repoName = repoObjName
	})

	When("repository object has empty default branch", func() {
		BeforeEach(func(ctx context.Context) {
			r := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: repoObjName, Namespace: namespace},
				Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: repoName},
			}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()
		})
		It("should update default branch and requeue", func(ctx context.Context) {
			ghRepo := &github.Repository{DefaultBranch: github.String("staging")}

			r := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKey{Name: repoObjName, Namespace: namespace}, r)).To(Succeed())
			result, err := act.NewSyncDefaultBranchAction(ghRepo).Execute(ctx, k, r)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeNil())

			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			Expect(rr.Status.DefaultBranch).To(Equal(ghRepo.GetDefaultBranch()))
			Expect(rr.Status.GetStaleCondition()).To(BeTrueDueTo(DefaultBranchOutOfSync))
		})
	})
	When("repository object has different default branch", func() {
		BeforeEach(func(ctx context.Context) {
			r := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: repoName},
				Status:     GitHubRepositoryStatus{DefaultBranch: strings.RandomHash(7)},
			}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()
		})
		It("should update default branch and requeue", func(ctx context.Context) {
			ghRepo := &github.Repository{DefaultBranch: github.String("staging")}
			r := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKey{Name: repoObjName, Namespace: namespace}, r)).To(Succeed())
			result, err := act.NewSyncDefaultBranchAction(ghRepo).Execute(ctx, k, r)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeNil())

			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			Expect(rr.Status.DefaultBranch).To(Equal(ghRepo.GetDefaultBranch()))
			Expect(rr.Status.GetStaleCondition()).To(BeTrueDueTo(DefaultBranchOutOfSync))
		})
	})
	When("repository object has correct default branch", func() {
		var ghRepo *github.Repository
		BeforeEach(func(ctx context.Context) {
			ghRepo := &github.Repository{DefaultBranch: github.String("staging")}
			r := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: repoName},
				Status:     GitHubRepositoryStatus{DefaultBranch: ghRepo.GetDefaultBranch()},
			}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()
		})
		It("should continue", func(ctx context.Context) {
			r := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKey{Name: repoObjName, Namespace: namespace}, r)).To(Succeed())
			result, err := act.NewSyncDefaultBranchAction(ghRepo).Execute(ctx, k, r)
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			Expect(rr.Status.DefaultBranch).To(Equal(ghRepo.GetDefaultBranch()))
			Expect(rr.Status.GetStaleCondition()).To(BeNil())
		})
	})
})
