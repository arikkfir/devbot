package github_test

import (
	"context"
	. "github.com/arikkfir/devbot/backend/api/v1"
	act "github.com/arikkfir/devbot/backend/internal/controllers/github"
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

var _ = Describe("NewFetchGitHubRepositoryAction", func() {
	const refreshInterval = 30 * time.Second
	var namespace, repoObjName string
	BeforeEach(func(ctx context.Context) {
		namespace = "default"
		repoObjName = strings.RandomHash(7)
	})

	When("repository owner is missing", func() {
		var k client.Client
		BeforeEach(func(ctx context.Context) {
			r := &GitHubRepository{ObjectMeta: metav1.ObjectMeta{Name: repoObjName, Namespace: namespace}}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()
		})
		It("should set conditions and abort", func(ctx context.Context) {
			r := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: repoObjName}, r)).To(Succeed())
			result, err := act.NewFetchGitHubRepositoryAction(refreshInterval, nil, nil).Execute(ctx, k, r)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&ctrl.Result{}))

			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			Expect(rr.Status.GetInvalidCondition()).To(BeTrueDueTo(RepositoryOwnerMissing))
			Expect(rr.Status.GetStaleCondition()).To(BeUnknownDueTo(Invalid))
			Expect(rr.Status.GetUnauthenticatedCondition()).To(BeTrueDueTo(Invalid))
		})
	})

	When("repository name is missing", func() {
		var k client.Client
		BeforeEach(func(ctx context.Context) {
			r := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: repoObjName, Namespace: namespace},
				Spec:       GitHubRepositorySpec{Owner: GitHubOwner},
			}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()
		})
		It("should set conditions and abort", func(ctx context.Context) {
			r := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: repoObjName}, r)).To(Succeed())
			result, err := act.NewFetchGitHubRepositoryAction(refreshInterval, nil, nil).Execute(ctx, k, r)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&ctrl.Result{}))

			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			Expect(r.Status.GetInvalidCondition()).To(BeTrueDueTo(RepositoryNameMissing))
			Expect(r.Status.GetStaleCondition()).To(BeUnknownDueTo(Invalid))
			Expect(r.Status.GetUnauthenticatedCondition()).To(BeTrueDueTo(Invalid))
		})
	})

	When("github repository config is valid", func() {
		var k client.Client

		When("repository does not exist", func() {
			var gh *github.Client
			BeforeEach(func(ctx context.Context) {
				r := &GitHubRepository{
					ObjectMeta: metav1.ObjectMeta{Name: repoObjName, Namespace: namespace},
					Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: strings.RandomHash(7)},
				}
				k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()
				gh = github.NewClient(mock.NewMockedHTTPClient(
					mock.WithRequestMatchHandler(
						mock.GetReposByOwnerByRepo,
						http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) }),
					),
				))
			})
			It("should set conditions and continue trying", func(ctx context.Context) {
				r := &GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: repoObjName}, r)).To(Succeed())
				var ghRepo *github.Repository
				result, err := act.NewFetchGitHubRepositoryAction(refreshInterval, gh, &ghRepo).Execute(ctx, k, r)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(&ctrl.Result{RequeueAfter: refreshInterval}))

				rr := &GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
				Expect(r.Status.GetInvalidCondition()).To(BeNil())
				Expect(r.Status.GetStaleCondition()).To(BeTrueDueTo(RepositoryNotFound))
				Expect(r.Status.GetUnauthenticatedCondition()).To(BeNil())
			})
		})

		When("github connection fails", func() {
			var gh *github.Client
			BeforeEach(func(ctx context.Context) {
				r := &GitHubRepository{
					ObjectMeta: metav1.ObjectMeta{Name: repoObjName, Namespace: "default"},
					Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: strings.RandomHash(7)},
				}
				k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()
				gh = github.NewClient(mock.NewMockedHTTPClient(
					mock.WithRequestMatchHandler(
						mock.GetReposByOwnerByRepo,
						http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }),
					),
				))
			})
			It("should set conditions and continue trying", func(ctx context.Context) {
				r := &GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: repoObjName}, r)).To(Succeed())
				var ghRepo *github.Repository
				result, err := act.NewFetchGitHubRepositoryAction(refreshInterval, gh, &ghRepo).Execute(ctx, k, r)

				rr := &GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(&ctrl.Result{RequeueAfter: refreshInterval}))
				Expect(rr.Status.GetInvalidCondition()).To(BeNil())
				Expect(rr.Status.GetStaleCondition()).To(BeUnknownDueTo(GitHubAPIFailed))
				Expect(rr.Status.GetUnauthenticatedCondition()).To(BeNil())
			})
		})

		When("github repository found", func() {
			var gh *github.Client
			var ghRepo github.Repository
			BeforeEach(func(ctx context.Context) {
				ghRepo = github.Repository{Name: github.String(strings.RandomHash(7))}
				gh = github.NewClient(mock.NewMockedHTTPClient(
					mock.WithRequestMatch(mock.GetReposByOwnerByRepo, ghRepo),
				))
				r := &GitHubRepository{
					ObjectMeta: metav1.ObjectMeta{Name: repoObjName, Namespace: namespace},
					Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: ghRepo.GetName()},
				}
				k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()
			})
			It("should store repository reference and continue", func(ctx context.Context) {
				r := &GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: repoObjName}, r)).To(Succeed())

				var fetchedGhRepo *github.Repository
				result, err := act.NewFetchGitHubRepositoryAction(refreshInterval, gh, &fetchedGhRepo).Execute(ctx, k, r)

				rr := &GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())
				Expect(*fetchedGhRepo).To(Equal(ghRepo))
				Expect(r.Status.GetInvalidCondition()).To(BeNil())
				Expect(r.Status.GetStaleCondition()).To(BeNil())
				Expect(r.Status.GetUnauthenticatedCondition()).To(BeNil())
			})
		})
	})
})
