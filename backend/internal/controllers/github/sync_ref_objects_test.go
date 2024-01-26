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
	"io"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("NewSyncGitHubRepositoryRefObjectsAction", func() {
	const namespace = "default"
	var k client.Client
	var r *GitHubRepository
	var repoMeta metav1.ObjectMeta
	var mainBranch *github.Branch
	var mainRef, staleRef *GitHubRepositoryRef

	BeforeEach(func(ctx context.Context) {
		repoMeta = metav1.ObjectMeta{Name: strings.RandomHash(7), Namespace: namespace}
		r = &GitHubRepository{
			ObjectMeta: repoMeta,
			Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: strings.RandomHash(7)},
		}
		mainBranch = &github.Branch{
			Name:   github.String("main"),
			Commit: &github.RepositoryCommit{SHA: github.String(strings.RandomHash(40))},
		}
		mainRef = &GitHubRepositoryRef{
			ObjectMeta: metav1.ObjectMeta{Name: "main", Namespace: namespace},
			Spec:       GitHubRepositoryRefSpec{Ref: "main"},
			Status: GitHubRepositoryRefStatus{
				RepositoryOwner: r.Spec.Owner,
				RepositoryName:  r.Spec.Name,
				CommitSHA:       mainBranch.GetCommit().GetSHA(),
			},
		}
		staleRef = &GitHubRepositoryRef{
			ObjectMeta: metav1.ObjectMeta{Name: "stale", Namespace: namespace},
			Spec:       GitHubRepositoryRefSpec{Ref: "stale"},
			Status: GitHubRepositoryRefStatus{
				RepositoryOwner: r.Spec.Owner,
				RepositoryName:  r.Spec.Name,
				CommitSHA:       strings.RandomHash(40),
			},
		}
	})

	When("there are refs without corresponding branches", func() {
		It("should delete stale refs", func(ctx context.Context) {
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r, mainRef, staleRef).WithStatusSubresource(r, mainRef, staleRef).Build()
			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			refs := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs)).To(Succeed())
			action := act.NewSyncGitHubRepositoryRefObjectsAction([]*github.Branch{mainBranch}, refs)
			Expect(action.Execute(ctx, k, rr)).To(BeNil())

			rrr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rrr)).To(Succeed())
			Expect(rrr.Status.GetStaleCondition()).To(BeNil())
			refs2 := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs2)).To(Succeed())
			Expect(refs2.Items).To(HaveLen(1))
			Expect(refs2.Items[0].Name).To(Equal(mainRef.Name))
		})
		It("should ignore stale refs that were already deleted", func(ctx context.Context) {
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r, mainRef).WithStatusSubresource(r, mainRef).Build()
			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			refs := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs)).To(Succeed())
			action := act.NewSyncGitHubRepositoryRefObjectsAction([]*github.Branch{mainBranch}, refs)
			Expect(action.Execute(ctx, k, rr)).To(BeNil())

			rrr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rrr)).To(Succeed())
			Expect(rrr.Status.GetStaleCondition()).To(BeNil())
			refs2 := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs2)).To(Succeed())
			Expect(refs2.Items).To(HaveLen(1))
			Expect(refs2.Items[0].Name).To(Equal(mainRef.Name))
		})
		It("should ignore deletion conflicts and requeue", func(ctx context.Context) {
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r, mainRef, staleRef).WithStatusSubresource(r, mainRef, staleRef).
				WithInterceptorFuncs(interceptor.Funcs{
					Delete: func(ctx context.Context, c client.WithWatch, o client.Object, opts ...client.DeleteOption) error {
						if o.GetNamespace() == staleRef.Namespace && o.GetName() == staleRef.Name {
							return apierrors.NewConflict(schema.GroupResource{}, o.GetName(), io.EOF)
						}
						return c.Delete(ctx, o, opts...)
					},
				}).
				Build()
			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			refs := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs)).To(Succeed())
			action := act.NewSyncGitHubRepositoryRefObjectsAction([]*github.Branch{mainBranch}, refs)
			Expect(action.Execute(ctx, k, rr)).To(Equal(&ctrl.Result{Requeue: true}))

			rrr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rrr)).To(Succeed())
			Expect(rrr.Status.GetStaleCondition()).To(BeTrueDueTo(BranchesOutOfSync))
			refs2 := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs2)).To(Succeed())
			Expect(refs2.Items).To(HaveLen(2))
		})
		It("should retry if ref deletion fails", func(ctx context.Context) {
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r, mainRef, staleRef).WithStatusSubresource(r, mainRef, staleRef).
				WithInterceptorFuncs(interceptor.Funcs{
					Delete: func(ctx context.Context, c client.WithWatch, o client.Object, opts ...client.DeleteOption) error {
						if o.GetNamespace() == staleRef.Namespace && o.GetName() == staleRef.Name {
							return apierrors.NewInternalError(io.EOF)
						}
						return c.Delete(ctx, o, opts...)
					},
				}).
				Build()
			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			refs := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs)).To(Succeed())
			action := act.NewSyncGitHubRepositoryRefObjectsAction([]*github.Branch{mainBranch}, refs)
			Expect(action.Execute(ctx, k, rr)).To(Equal(&ctrl.Result{Requeue: true}))

			rrr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rrr)).To(Succeed())
			Expect(rrr.Status.GetStaleCondition()).To(BeTrueDueTo(BranchesOutOfSync))
			refs2 := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs2)).To(Succeed())
			Expect(refs2.Items).To(HaveLen(2))
		})
	})

	When("there are refs out of sync", func() {
		It("should sync ref owner, name and commit SHA", func(ctx context.Context) {
			mainRef.Status.RepositoryOwner = strings.RandomHash(7)
			mainRef.Status.RepositoryName = strings.RandomHash(7)
			mainRef.Status.CommitSHA = strings.RandomHash(7)
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r, mainRef).WithStatusSubresource(r, mainRef).Build()
			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			refs := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs)).To(Succeed())
			action := act.NewSyncGitHubRepositoryRefObjectsAction([]*github.Branch{mainBranch}, refs)
			Expect(action.Execute(ctx, k, rr)).To(BeNil())

			rrr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rrr)).To(Succeed())
			Expect(rrr.Status.GetStaleCondition()).To(BeNil())
			refs2 := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs2)).To(Succeed())
			Expect(refs2.Items).To(HaveLen(1))
			Expect(refs2.Items[0].Name).To(Equal(mainRef.Name))
			Expect(refs2.Items[0].Status.RepositoryOwner).To(Equal(r.Spec.Owner))
			Expect(refs2.Items[0].Status.RepositoryName).To(Equal(r.Spec.Name))
			Expect(refs2.Items[0].Status.CommitSHA).To(Equal(mainBranch.GetCommit().GetSHA()))
		})
		It("should ignore refs that were already deleted", func(ctx context.Context) {
			mainRef.Status.RepositoryOwner = strings.RandomHash(7)
			mainRef.Status.RepositoryName = strings.RandomHash(7)
			mainRef.Status.CommitSHA = strings.RandomHash(7)
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()
			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			refs := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs)).To(Succeed())
			action := act.NewSyncGitHubRepositoryRefObjectsAction([]*github.Branch{mainBranch}, refs)
			Expect(action.Execute(ctx, k, rr)).To(BeNil())

			rrr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rrr)).To(Succeed())
			Expect(rrr.Status.GetStaleCondition()).To(BeNil())
			refs2 := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs2)).To(Succeed())
			Expect(refs2.Items).To(BeEmpty())
		})
		It("should ignore sync conflicts and requeue", func(ctx context.Context) {
			mainRef.Status.RepositoryOwner = strings.RandomHash(7)
			mainRef.Status.RepositoryName = strings.RandomHash(7)
			mainRef.Status.CommitSHA = strings.RandomHash(7)
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r, mainRef).WithStatusSubresource(r, mainRef).
				WithInterceptorFuncs(interceptor.Funcs{
					SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, o client.Object, opts ...client.SubResourceUpdateOption) error {
						if subResourceName == "status" && o.GetNamespace() == mainRef.Namespace && o.GetName() == mainRef.Name {
							return apierrors.NewConflict(schema.GroupResource{}, o.GetName(), io.EOF)
						}
						return c.Status().Update(ctx, o, opts...)
					},
				}).
				Build()
			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			refs := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs)).To(Succeed())
			action := act.NewSyncGitHubRepositoryRefObjectsAction([]*github.Branch{mainBranch}, refs)
			Expect(action.Execute(ctx, k, rr)).To(Equal(&ctrl.Result{Requeue: true}))

			rrr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rrr)).To(Succeed())
			Expect(rrr.Status.GetStaleCondition()).To(BeTrueDueTo(BranchesOutOfSync))
			refs2 := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs2)).To(Succeed())
			Expect(refs2.Items).To(HaveLen(1))
			Expect(refs2.Items[0].Name).To(Equal(mainRef.Name))
			Expect(refs2.Items[0].Status.RepositoryOwner).To(Equal(mainRef.Status.RepositoryOwner))
			Expect(refs2.Items[0].Status.RepositoryName).To(Equal(mainRef.Status.RepositoryName))
			Expect(refs2.Items[0].Status.CommitSHA).To(Equal(mainRef.Status.CommitSHA))
		})
		It("should retry if ref sync fails", func(ctx context.Context) {
			mainRef.Status.RepositoryOwner = strings.RandomHash(7)
			mainRef.Status.RepositoryName = strings.RandomHash(7)
			mainRef.Status.CommitSHA = strings.RandomHash(7)
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r, mainRef).WithStatusSubresource(r, mainRef).
				WithInterceptorFuncs(interceptor.Funcs{
					SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, o client.Object, opts ...client.SubResourceUpdateOption) error {
						if subResourceName == "status" && o.GetNamespace() == mainRef.Namespace && o.GetName() == mainRef.Name {
							return apierrors.NewInternalError(io.EOF)
						}
						return c.Status().Update(ctx, o, opts...)
					},
				}).
				Build()
			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			refs := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs)).To(Succeed())
			action := act.NewSyncGitHubRepositoryRefObjectsAction([]*github.Branch{mainBranch}, refs)
			Expect(action.Execute(ctx, k, rr)).To(Equal(&ctrl.Result{Requeue: true}))

			rrr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rrr)).To(Succeed())
			Expect(rrr.Status.GetStaleCondition()).To(BeTrueDueTo(BranchesOutOfSync))
			refs2 := &GitHubRepositoryRefList{}
			Expect(k.List(ctx, refs2)).To(Succeed())
			Expect(refs2.Items).To(HaveLen(1))
			Expect(refs2.Items[0].Name).To(Equal(mainRef.Name))
			Expect(refs2.Items[0].Status.RepositoryOwner).To(Equal(mainRef.Status.RepositoryOwner))
			Expect(refs2.Items[0].Status.RepositoryName).To(Equal(mainRef.Status.RepositoryName))
			Expect(refs2.Items[0].Status.CommitSHA).To(Equal(mainRef.Status.CommitSHA))
		})
	})
})
