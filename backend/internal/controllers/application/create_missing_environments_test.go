package application_test

import (
	"context"
	. "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/controllers/application"
	strings2 "github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
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
	"slices"
	"strings"
)

var _ = Describe("NewCreateMissingEnvironmentsAction", func() {
	var namespace, appName string

	BeforeEach(func() { namespace, appName = "default", strings2.RandomHash(7) })

	It("should be marked as stale when repositories are not found", func(ctx context.Context) {
		app := &Application{
			ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: namespace},
			Spec: ApplicationSpec{
				Repositories: []ApplicationSpecRepository{
					{
						RepositoryReferenceWithOptionalNamespace: RepositoryReferenceWithOptionalNamespace{
							APIVersion: GitHubRepositoryGVK.GroupVersion().String(),
							Kind:       GitHubRepositoryGVK.Kind,
							Name:       "repo",
							Namespace:  namespace,
						},
						MissingBranchStrategy: MissingBranchStrategyUseDefaultBranch,
					},
				},
			},
		}
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app).WithStatusSubresource(app).Build()

		a := &Application{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(app), a)).To(Succeed())
		result, err := application.NewCreateMissingEnvironmentsAction(&EnvironmentList{}).Execute(ctx, k, a)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(&ctrl.Result{Requeue: true}))
		Expect(a.Status.GetStaleCondition()).To(BeTrueDueTo(RepositoryNotFound))
	})
	It("should be marked as stale when repositories are not accessible", func(ctx context.Context) {
		repo := &GitHubRepository{
			ObjectMeta: metav1.ObjectMeta{Name: strings2.RandomHash(7), Namespace: namespace},
			Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: "repo"},
		}
		app := &Application{
			ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: namespace},
			Spec: ApplicationSpec{
				Repositories: []ApplicationSpecRepository{
					{
						RepositoryReferenceWithOptionalNamespace: RepositoryReferenceWithOptionalNamespace{
							APIVersion: GitHubRepositoryGVK.GroupVersion().String(),
							Kind:       GitHubRepositoryGVK.Kind,
							Name:       repo.Name,
							Namespace:  repo.Namespace,
						},
						MissingBranchStrategy: MissingBranchStrategyUseDefaultBranch,
					},
				},
			},
		}
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app, repo).WithStatusSubresource(app, repo).
			WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if key == client.ObjectKeyFromObject(repo) {
						return apierrors.NewForbidden(schema.GroupResource{}, repo.Name, io.EOF)
					}
					return c.Get(ctx, key, obj, opts...)
				},
			}).
			Build()

		a := &Application{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(app), a)).To(Succeed())
		result, err := application.NewCreateMissingEnvironmentsAction(&EnvironmentList{}).Execute(ctx, k, a)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(&ctrl.Result{Requeue: true}))
		Expect(a.Status.GetStaleCondition()).To(BeUnknownDueTo(RepositoryNotAccessible))
	})
	It("should be marked as possibly-stale when repositories cannot be fetched", func(ctx context.Context) {
		repo := &GitHubRepository{
			ObjectMeta: metav1.ObjectMeta{Name: strings2.RandomHash(7), Namespace: namespace},
			Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: "repo"},
		}
		app := &Application{
			ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: namespace},
			Spec: ApplicationSpec{
				Repositories: []ApplicationSpecRepository{
					{
						RepositoryReferenceWithOptionalNamespace: RepositoryReferenceWithOptionalNamespace{
							APIVersion: GitHubRepositoryGVK.GroupVersion().String(),
							Kind:       GitHubRepositoryGVK.Kind,
							Name:       repo.Name,
							Namespace:  repo.Namespace,
						},
						MissingBranchStrategy: MissingBranchStrategyUseDefaultBranch,
					},
				},
			},
		}
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app, repo).WithStatusSubresource(app, repo).
			WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if key == client.ObjectKeyFromObject(repo) {
						return apierrors.NewInternalError(io.EOF)
					}
					return c.Get(ctx, key, obj, opts...)
				},
			}).
			Build()

		a := &Application{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(app), a)).To(Succeed())
		result, err := application.NewCreateMissingEnvironmentsAction(&EnvironmentList{}).Execute(ctx, k, a)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(&ctrl.Result{Requeue: true}))
		Expect(a.Status.GetStaleCondition()).To(BeUnknownDueTo(InternalError))
	})
	It("should be marked as possibly-stale when refs cannot be fetched", func(ctx context.Context) {
		repo := &GitHubRepository{
			ObjectMeta: metav1.ObjectMeta{Name: strings2.RandomHash(7), Namespace: namespace},
			Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: "repo"},
		}
		mainRef := &GitHubRepositoryRef{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "main",
				Namespace:       repo.Namespace,
				OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(repo, GitHubRepositoryGVK)},
			},
			Spec: GitHubRepositoryRefSpec{Ref: "main"},
		}
		app := &Application{
			ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: namespace},
			Spec: ApplicationSpec{
				Repositories: []ApplicationSpecRepository{
					{
						RepositoryReferenceWithOptionalNamespace: RepositoryReferenceWithOptionalNamespace{
							APIVersion: GitHubRepositoryGVK.GroupVersion().String(),
							Kind:       GitHubRepositoryGVK.Kind,
							Name:       repo.Name,
							Namespace:  repo.Namespace,
						},
						MissingBranchStrategy: MissingBranchStrategyUseDefaultBranch,
					},
				},
			},
		}
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app, repo, mainRef).WithStatusSubresource(app, repo, mainRef).
			WithInterceptorFuncs(interceptor.Funcs{
				List: func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
					switch list.(type) {
					case *GitHubRepositoryRefList:
						return apierrors.NewInternalError(io.EOF)
					default:
						return client.List(ctx, list, opts...)
					}
				},
			}).
			Build()

		a := &Application{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(app), a)).To(Succeed())
		result, err := application.NewCreateMissingEnvironmentsAction(&EnvironmentList{}).Execute(ctx, k, a)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(&ctrl.Result{Requeue: true}))
		Expect(a.Status.GetStaleCondition()).To(BeUnknownDueTo(InternalError))
	})
	It("should be marked as possibly-stale when envs cannot be created", func(ctx context.Context) {
		repo := &GitHubRepository{
			ObjectMeta: metav1.ObjectMeta{Name: strings2.RandomHash(7), Namespace: namespace},
			Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: "repo"},
		}
		mainRef := &GitHubRepositoryRef{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "main",
				Namespace:       repo.Namespace,
				OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(repo, GitHubRepositoryGVK)},
			},
			Spec: GitHubRepositoryRefSpec{Ref: "main"},
		}
		app := &Application{
			ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: namespace},
			Spec: ApplicationSpec{
				Repositories: []ApplicationSpecRepository{
					{
						RepositoryReferenceWithOptionalNamespace: RepositoryReferenceWithOptionalNamespace{
							APIVersion: GitHubRepositoryGVK.GroupVersion().String(),
							Kind:       GitHubRepositoryGVK.Kind,
							Name:       repo.Name,
							Namespace:  repo.Namespace,
						},
						MissingBranchStrategy: MissingBranchStrategyUseDefaultBranch,
					},
				},
			},
		}
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app, repo, mainRef).WithStatusSubresource(app, repo, mainRef).
			WithInterceptorFuncs(interceptor.Funcs{
				Create: func(ctx context.Context, c client.WithWatch, o client.Object, opts ...client.CreateOption) error {
					if _, ok := o.(*Environment); ok {
						return apierrors.NewInternalError(io.EOF)
					}
					return c.Create(ctx, o, opts...)
				},
			}).
			Build()

		a := &Application{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(app), a)).To(Succeed())
		result, err := application.NewCreateMissingEnvironmentsAction(&EnvironmentList{}).Execute(ctx, k, a)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(&ctrl.Result{Requeue: true}))
		Expect(a.Status.GetStaleCondition()).To(BeUnknownDueTo(InternalError))
	})
	It("should create distinct environments from refs of participating repositories", func(ctx context.Context) {
		repo := &GitHubRepository{
			ObjectMeta: metav1.ObjectMeta{Name: strings2.RandomHash(7), Namespace: namespace},
			Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: "repo"},
		}
		mainRef := &GitHubRepositoryRef{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "main",
				Namespace:       repo.Namespace,
				OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(repo, GitHubRepositoryGVK)},
			},
			Spec: GitHubRepositoryRefSpec{Ref: "main"},
		}
		feature1Ref := &GitHubRepositoryRef{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "feature1",
				Namespace:       repo.Namespace,
				OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(repo, GitHubRepositoryGVK)},
			},
			Spec: GitHubRepositoryRefSpec{Ref: "feature1"},
		}
		app := &Application{
			ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: namespace},
			Spec: ApplicationSpec{
				Repositories: []ApplicationSpecRepository{
					{
						RepositoryReferenceWithOptionalNamespace: RepositoryReferenceWithOptionalNamespace{
							APIVersion: GitHubRepositoryGVK.GroupVersion().String(),
							Kind:       GitHubRepositoryGVK.Kind,
							Name:       repo.Name,
							Namespace:  repo.Namespace,
						},
						MissingBranchStrategy: MissingBranchStrategyUseDefaultBranch,
					},
				},
			},
		}
		k := fake.NewClientBuilder().WithScheme(scheme).
			WithIndex(&GitHubRepositoryRef{}, k8s.OwnershipIndexField, k8s.IndexGetOwnerReferencesOf).
			WithObjects(app, repo, mainRef, feature1Ref).
			WithStatusSubresource(app, repo, mainRef, feature1Ref).
			Build()

		a := &Application{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(app), a)).To(Succeed())
		result, err := application.NewCreateMissingEnvironmentsAction(&EnvironmentList{}).Execute(ctx, k, a)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeNil())
		Expect(a.Status.GetStaleCondition()).To(BeNil())

		envList := &EnvironmentList{}
		Expect(k.List(ctx, envList)).To(Succeed())
		slices.SortFunc(envList.Items, func(i, j Environment) int { return strings.Compare(i.Spec.PreferredBranch, j.Spec.PreferredBranch) })
		Expect(envList.Items).To(HaveLen(2))
		Expect(envList.Items[0].Spec.PreferredBranch).To(Equal("feature1"))
		Expect(envList.Items[0].OwnerReferences).To(HaveLen(1))
		Expect(envList.Items[0].OwnerReferences[0].APIVersion).To(Equal(ApplicationGVK.GroupVersion().String()))
		Expect(envList.Items[0].OwnerReferences[0].Kind).To(Equal(ApplicationGVK.Kind))
		Expect(envList.Items[0].OwnerReferences[0].Name).To(Equal(a.Name))
		Expect(envList.Items[0].OwnerReferences[0].UID).To(Equal(a.UID))
		Expect(envList.Items[0].OwnerReferences[0].Controller).To(Equal(&[]bool{true}[0]))
		Expect(envList.Items[1].Spec.PreferredBranch).To(Equal("main"))
		Expect(envList.Items[1].OwnerReferences).To(HaveLen(1))
		Expect(envList.Items[1].OwnerReferences[0].APIVersion).To(Equal(ApplicationGVK.GroupVersion().String()))
		Expect(envList.Items[1].OwnerReferences[0].Kind).To(Equal(ApplicationGVK.Kind))
		Expect(envList.Items[1].OwnerReferences[0].Name).To(Equal(a.Name))
		Expect(envList.Items[1].OwnerReferences[0].UID).To(Equal(a.UID))
		Expect(envList.Items[1].OwnerReferences[0].Controller).To(Equal(&[]bool{true}[0]))
	})
})
