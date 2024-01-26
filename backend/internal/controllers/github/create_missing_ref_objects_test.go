package github_test

import (
	"cmp"
	"context"
	. "github.com/arikkfir/devbot/backend/api/v1"
	act "github.com/arikkfir/devbot/backend/internal/controllers/github"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"slices"
)

var _ = Describe("NewCreateMissingGitHubRepositoryRefObjectsAction", func() {
	var k client.Client
	When("no branches are given", func() {
		BeforeEach(func(ctx context.Context) {
			k = fake.NewClientBuilder().
				WithScheme(scheme).
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
						Fail("Create should not be called")
						return c.Create(ctx, obj, opts...)
					},
				}).
				Build()
		})
		When("no refs are given", func() {
			It("should do nothing and continue", func(ctx context.Context) {
				result, err := act.NewCreateMissingGitHubRepositoryRefObjectsAction(nil, nil).Execute(ctx, k, &GitHubRepository{})
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())
			})
		})
		When("1 refs is given", func() {
			It("should do nothing continue", func(ctx context.Context) {
				action := act.NewCreateMissingGitHubRepositoryRefObjectsAction(
					nil,
					&GitHubRepositoryRefList{
						Items: []GitHubRepositoryRef{
							{Spec: GitHubRepositoryRefSpec{Ref: "main"}},
						},
					},
				)
				result, err := action.Execute(ctx, k, &GitHubRepository{})
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())
			})
		})
		When("2 refs are given", func() {
			It("should do nothing continue", func(ctx context.Context) {
				action := act.NewCreateMissingGitHubRepositoryRefObjectsAction(
					nil,
					&GitHubRepositoryRefList{
						Items: []GitHubRepositoryRef{
							{Spec: GitHubRepositoryRefSpec{Ref: "main"}},
							{Spec: GitHubRepositoryRefSpec{Ref: "b1"}},
						},
					},
				)
				result, err := action.Execute(ctx, k, &GitHubRepository{})
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())
			})
		})
	})
	When("one branch is given", func() {
		var branches []*github.Branch
		BeforeEach(func(ctx context.Context) {
			branches = []*github.Branch{
				{Name: github.String("main")},
			}
		})
		When("no refs are given", func() {
			BeforeEach(func(ctx context.Context) { k = fake.NewClientBuilder().WithScheme(scheme).Build() })
			It("should create one ref object and requeue", func(ctx context.Context) {
				r := &GitHubRepository{
					ObjectMeta: metav1.ObjectMeta{Name: "r", Namespace: "default", UID: "abc"},
				}
				action := act.NewCreateMissingGitHubRepositoryRefObjectsAction(branches, &GitHubRepositoryRefList{})
				result, err := action.Execute(ctx, k, r)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(&ctrl.Result{Requeue: true}))

				refs := &GitHubRepositoryRefList{}
				Expect(k.List(ctx, refs)).To(Succeed())
				Expect(refs.Items).To(HaveLen(1))
				Expect(refs.Items[0].Namespace).To(Equal("default"))
				Expect(refs.Items[0].OwnerReferences).To(ConsistOf(metav1.OwnerReference{
					Name: r.Name, UID: r.UID, Controller: &[]bool{true}[0],
				}))
				Expect(refs.Items[0].Spec.Ref).To(Equal("main"))
			})
		})
		When("1 refs is given", func() {
			When("the ref is the same as the branch", func() {
				BeforeEach(func(ctx context.Context) {
					k = fake.NewClientBuilder().
						WithScheme(scheme).
						WithInterceptorFuncs(interceptor.Funcs{
							Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
								Fail("Create should not be called")
								return c.Create(ctx, obj, opts...)
							},
						}).
						Build()
				})
				It("should not create a ref object and continue", func(ctx context.Context) {
					action := act.NewCreateMissingGitHubRepositoryRefObjectsAction(branches, &GitHubRepositoryRefList{
						Items: []GitHubRepositoryRef{
							{Spec: GitHubRepositoryRefSpec{Ref: "main"}},
						},
					})
					result, err := action.Execute(ctx, k, &GitHubRepository{})
					Expect(err).To(BeNil())
					Expect(result).To(BeNil())
				})
			})
			When("the ref is different than the branch", func() {
				BeforeEach(func(ctx context.Context) { k = fake.NewClientBuilder().WithScheme(scheme).Build() })
				It("should create a ref object and requeue", func(ctx context.Context) {
					r := &GitHubRepository{
						ObjectMeta: metav1.ObjectMeta{Name: "r", Namespace: "default", UID: "abc"},
					}
					action := act.NewCreateMissingGitHubRepositoryRefObjectsAction(
						branches,
						&GitHubRepositoryRefList{
							Items: []GitHubRepositoryRef{{Spec: GitHubRepositoryRefSpec{Ref: "b1"}}},
						})
					result, err := action.Execute(ctx, k, r)
					Expect(err).To(BeNil())
					Expect(result).To(BeNil())

					refs := &GitHubRepositoryRefList{}
					Expect(k.List(ctx, refs)).To(Succeed())
					Expect(refs.Items).To(HaveLen(2))
					slices.SortFunc(refs.Items, func(i, j GitHubRepositoryRef) int { return cmp.Compare(i.Spec.Ref, j.Spec.Ref) })
					Expect(refs.Items[0].Spec.Ref).To(Equal("b1"))
					Expect(refs.Items[1].Name).ToNot(BeEmpty())
					Expect(refs.Items[1].Namespace).To(Equal(r.Namespace))
					Expect(refs.Items[1].Spec.Ref).To(Equal("main"))
					Expect(refs.Items[1].OwnerReferences).To(ConsistOf(metav1.OwnerReference{
						Name: r.Name, UID: r.UID, Controller: &[]bool{true}[0],
					}))
				})
			})
		})
	})
})
