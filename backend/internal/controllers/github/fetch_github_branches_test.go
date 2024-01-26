package github_test

import (
	"context"
	. "github.com/arikkfir/devbot/backend/api/v1"
	t "github.com/arikkfir/devbot/backend/internal/controllers/github"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	"github.com/google/go-github/v56/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"time"
)

var _ = Describe("NewFetchGitHubBranchesAction", func() {
	const refreshInterval = 30 * time.Second
	var gh *github.Client
	var namespace, repoName string
	BeforeEach(func() { namespace, repoName = "test-ns", strings.RandomHash(7) })
	When("github client fails to list branches", func() {
		var k client.WithWatch
		var r *GitHubRepository
		BeforeEach(func(ctx context.Context) {
			gh = github.NewClient(mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposBranchesByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusInternalServerError)
					}),
				),
			))
			r = &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
				Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: repoName},
			}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()
		})
		It("should set stale condition and requeue", func(ctx context.Context) {
			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			result, err := t.NewFetchGitHubBranchesAction(refreshInterval, gh, nil).Execute(ctx, k, rr)
			Expect(err).To(BeNil())
			Expect(result).To(Equal(&ctrl.Result{RequeueAfter: refreshInterval}))

			rrr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rrr)).To(Succeed())
			Expect(rrr.Status.GetStaleCondition()).To(BeUnknownDueTo(GitHubAPIFailed))
		})
	})
	When("repository has no branches", func() {
		BeforeEach(func(ctx context.Context) {
			gh = github.NewClient(mock.NewMockedHTTPClient(
				mock.WithRequestMatch(mock.GetReposBranchesByOwnerByRepo, []github.Branch{}),
			))
		})
		It("should set target branches array to nil and continue", func(ctx context.Context) {
			r := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
				Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: repoName},
			}
			k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()
			var targetBranches []*github.Branch
			result, err := t.NewFetchGitHubBranchesAction(refreshInterval, gh, &targetBranches).Execute(ctx, k, r)
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
			Expect(targetBranches).To(BeNil())
		})
	})
	When("repository has multiple branches", func() {
		When("github client succeeds to list branches", func() {
			BeforeEach(func(ctx context.Context) {
				gh = github.NewClient(mock.NewMockedHTTPClient(
					mock.WithRequestMatch(
						mock.GetReposBranchesByOwnerByRepo,
						[]github.Branch{
							{Name: github.String("main")},
							{Name: github.String("b1")},
							{Name: github.String("b2")},
						},
					),
				))
			})
			It("should return correct branches and continue", func(ctx context.Context) {
				r := &GitHubRepository{
					ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
					Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: repoName},
				}
				k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()
				var branches []*github.Branch
				result, err := t.NewFetchGitHubBranchesAction(refreshInterval, gh, &branches).Execute(ctx, k, r)
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())
				Expect(branches).To(HaveLen(3))
				Expect(branches).To(ConsistOf(BranchHasName("main"), BranchHasName("b1"), BranchHasName("b2")))
			})
		})
	})
})
