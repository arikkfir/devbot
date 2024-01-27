package deployment_test

import (
	"context"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/controllers/deployment"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
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

var _ = Describe("NewPrepareCloneDirectoryAction", func() {
	var namespace, deploymentName string
	BeforeEach(func(ctx context.Context) { namespace, deploymentName = "default", strings.RandomHash(7) })

	It("should set conditions and abort when repository is not found", func(ctx context.Context) {
		o := &apiv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespace},
			Spec: apiv1.DeploymentSpec{
				Repository: apiv1.NamespacedRepositoryReference{
					APIVersion: apiv1.GitHubRepositoryGVK.GroupVersion().String(),
					Kind:       apiv1.GitHubRepositoryGVK.Kind,
					Name:       "repo",
					Namespace:  namespace,
				},
				Branch: "main",
			},
		}
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(o).WithStatusSubresource(o).Build()
		oo := &apiv1.Deployment{}
		var gitURL string
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
		result, err := deployment.NewPrepareCloneDirectoryAction(&gitURL).Execute(ctx, k, oo)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(&ctrl.Result{Requeue: true}))

		ooo := &apiv1.Deployment{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), ooo)).To(Succeed())
		Expect(ooo.Status.GetInvalidCondition()).To(BeTrueDueTo(apiv1.RepositoryNotFound))
		Expect(ooo.Status.GetStaleCondition()).To(BeUnknownDueTo(apiv1.Invalid))
	})

	It("should set conditions and abort when repository is not accessible", func(ctx context.Context) {
		repo := &apiv1.GitHubRepository{
			ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: namespace},
			Spec: apiv1.GitHubRepositorySpec{
				Owner: GitHubOwner,
				Name:  strings.RandomHash(7),
			},
		}
		o := &apiv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespace},
			Spec: apiv1.DeploymentSpec{
				Repository: apiv1.NamespacedRepositoryReference{
					APIVersion: apiv1.GitHubRepositoryGVK.GroupVersion().String(),
					Kind:       apiv1.GitHubRepositoryGVK.Kind,
					Name:       "repo",
					Namespace:  namespace,
				},
				Branch: "main",
			},
		}
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(o, repo).WithStatusSubresource(o, repo).
			WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if key == client.ObjectKeyFromObject(repo) {
						return apierrors.NewForbidden(schema.GroupResource{}, repo.Name, io.EOF)
					}
					return c.Get(ctx, key, obj, opts...)
				},
			}).
			Build()
		oo := &apiv1.Deployment{}
		var gitURL string
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
		result, err := deployment.NewPrepareCloneDirectoryAction(&gitURL).Execute(ctx, k, oo)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(&ctrl.Result{Requeue: true}))

		ooo := &apiv1.Deployment{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), ooo)).To(Succeed())
		Expect(ooo.Status.GetInvalidCondition()).To(BeTrueDueTo(apiv1.RepositoryNotAccessible))
		Expect(ooo.Status.GetStaleCondition()).To(BeUnknownDueTo(apiv1.Invalid))
	})

	It("should set conditions and abort when repository cannot be fetched", func(ctx context.Context) {
		repo := &apiv1.GitHubRepository{
			ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: namespace},
			Spec: apiv1.GitHubRepositorySpec{
				Owner: GitHubOwner,
				Name:  strings.RandomHash(7),
			},
		}
		o := &apiv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespace},
			Spec: apiv1.DeploymentSpec{
				Repository: apiv1.NamespacedRepositoryReference{
					APIVersion: apiv1.GitHubRepositoryGVK.GroupVersion().String(),
					Kind:       apiv1.GitHubRepositoryGVK.Kind,
					Name:       "repo",
					Namespace:  namespace,
				},
				Branch: "main",
			},
		}
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(o, repo).WithStatusSubresource(o, repo).
			WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if key == client.ObjectKeyFromObject(repo) {
						return apierrors.NewInternalError(io.EOF)
					}
					return c.Get(ctx, key, obj, opts...)
				},
			}).
			Build()
		oo := &apiv1.Deployment{}
		var gitURL string
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
		result, err := deployment.NewPrepareCloneDirectoryAction(&gitURL).Execute(ctx, k, oo)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(&ctrl.Result{Requeue: true}))

		ooo := &apiv1.Deployment{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), ooo)).To(Succeed())
		Expect(ooo.Status.GetInvalidCondition()).To(BeNil())
		Expect(ooo.Status.GetStaleCondition()).To(BeUnknownDueTo(apiv1.InternalError))
	})

	It("should set target clone path and Git URL", func(ctx context.Context) {
		repo := &apiv1.GitHubRepository{
			ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: namespace},
			Spec: apiv1.GitHubRepositorySpec{
				Owner: GitHubOwner,
				Name:  strings.RandomHash(7),
			},
		}
		o := &apiv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespace},
			Spec: apiv1.DeploymentSpec{
				Repository: apiv1.NamespacedRepositoryReference{
					APIVersion: apiv1.GitHubRepositoryGVK.GroupVersion().String(),
					Kind:       apiv1.GitHubRepositoryGVK.Kind,
					Name:       "repo",
					Namespace:  namespace,
				},
				Branch: "main",
			},
		}
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(o, repo).WithStatusSubresource(o, repo).Build()
		oo := &apiv1.Deployment{}
		var gitURL string
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
		result, err := deployment.NewPrepareCloneDirectoryAction(&gitURL).Execute(ctx, k, oo)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(&ctrl.Result{Requeue: true}))
		Expect(gitURL).To(BeEmpty())

		ooo := &apiv1.Deployment{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), ooo)).To(Succeed())
		Expect(ooo.Status.ClonePath).To(HavePrefix(fmt.Sprintf("/data/%s/%s/", repo.Namespace, repo.Name)))
		Expect(ooo.Status.ClonePath).ToNot(Equal(fmt.Sprintf("/data/%s/%s/", repo.Namespace, repo.Name)))
		Expect(ooo.Status.GetInvalidCondition()).To(BeNil())
		Expect(ooo.Status.GetStaleCondition()).To(BeNil())

		oo = &apiv1.Deployment{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
		result, err = deployment.NewPrepareCloneDirectoryAction(&gitURL).Execute(ctx, k, oo)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeNil())
		Expect(gitURL).To(Equal(fmt.Sprintf("https://github.com/%s/%s", repo.Spec.Owner, repo.Spec.Name)))

		ooo = &apiv1.Deployment{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), ooo)).To(Succeed())
		Expect(ooo.Status.ClonePath).To(HavePrefix(fmt.Sprintf("/data/%s/%s/", repo.Namespace, repo.Name)))
		Expect(ooo.Status.ClonePath).ToNot(Equal(fmt.Sprintf("/data/%s/%s/", repo.Namespace, repo.Name)))
		Expect(ooo.Status.GetInvalidCondition()).To(BeNil())
		Expect(ooo.Status.GetStaleCondition()).To(BeTrueDueTo(apiv1.CloneMissing))
	})
})
