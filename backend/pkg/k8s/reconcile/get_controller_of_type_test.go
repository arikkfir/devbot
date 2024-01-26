package reconcile_test

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"time"
)

var _ = Describe("NewGetControllerAction", func() {
	var k client.WithWatch
	When("controller reference is defined", func() {
		var namespace, controllerName, nonControllerName, refName string
		BeforeEach(func(ctx context.Context) {
			namespace = "default"
			controllerName = strings.RandomHash(7)
			nonControllerName = strings.RandomHash(7)
			refName = strings.RandomHash(7)
			controller := &apiv1.GitHubRepository{ObjectMeta: metav1.ObjectMeta{Name: controllerName, Namespace: namespace, UID: "1"}}
			nonController := &apiv1.GitHubRepository{ObjectMeta: metav1.ObjectMeta{Name: nonControllerName, Namespace: namespace, UID: "2"}}
			controllee := &apiv1.GitHubRepositoryRef{
				ObjectMeta: metav1.ObjectMeta{
					Name:      refName,
					Namespace: namespace,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(controller, apiv1.GitHubRepositoryGVK),
						*k8s.NewOwnerReference(nonController, apiv1.GitHubRepositoryGVK, &[]bool{false}[0]),
					},
				},
			}
			k = fake.NewClientBuilder().
				WithScheme(scheme).
				WithIndex(&apiv1.GitHubRepositoryRef{}, k8s.OwnershipIndexField, k8s.IndexGetOwnerReferencesOf).
				WithObjects(controller, nonController, controllee).
				WithStatusSubresource(controllee).
				Build()
		})
		When("controller exists", func() {
			It("should obtain controller and continue", func(ctx context.Context) {
				var expectedController = &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: controllerName}, expectedController)).To(Succeed())

				var ref = &apiv1.GitHubRepositoryRef{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: refName}, ref)).To(Succeed())

				var actualController = &apiv1.GitHubRepository{}
				result, err := reconcile.NewGetControllerAction(false, actualController).Execute(ctx, k, ref)
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())
				Expect(actualController).To(Equal(expectedController))
			})
		})
		When("controller does not exist", func() {
			BeforeEach(func(ctx context.Context) {
				k = interceptor.NewClient(k, interceptor.Funcs{
					Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if key.Name == controllerName {
							return apierrors.NewNotFound(schema.GroupResource{}, key.Name)
						}
						return c.Get(ctx, key, obj, opts...)
					},
				})
			})
			When("controller is required", func() {
				It("should set invalid status and stop", func(ctx context.Context) {
					var o = &apiv1.GitHubRepositoryRef{}
					Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: refName}, o)).To(Succeed())

					result, err := reconcile.NewGetControllerAction(false, &apiv1.GitHubRepository{}).Execute(ctx, k, o)
					Expect(err).To(BeNil())
					Expect(result).ToNot(BeNil())
					Expect(result.Requeue).To(BeFalse())
					Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

					o = &apiv1.GitHubRepositoryRef{}
					Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: refName}, o)).To(Succeed())
					Expect(o.Status.GetInvalidCondition()).To(BeTrueDueTo(apiv1.ControllerMissing))
				})
			})
			When("controller is optional", func() {
				It("should stop", func(ctx context.Context) {
					var o = &apiv1.GitHubRepositoryRef{}
					Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: refName}, o)).To(Succeed())

					var controller client.Object
					result, err := reconcile.NewGetControllerAction(true, controller).Execute(ctx, k, o)
					Expect(err).To(BeNil())
					Expect(result).To(BeNil())

					o = &apiv1.GitHubRepositoryRef{}
					Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: refName}, o)).To(Succeed())
					Expect(o.Status.IsValid()).To(BeTrue())
				})
			})
		})
		When("controller could not be fetched", func() {
			BeforeEach(func(ctx context.Context) {
				k = interceptor.NewClient(k, interceptor.Funcs{
					Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if key.Name == controllerName {
							return apierrors.NewInternalError(io.EOF)
						}
						return c.Get(ctx, key, obj, opts...)
					},
				})
			})
			DescribeTable("should set invalid status and requeue with error",
				func(ctx context.Context, optional bool) {
					var o = &apiv1.GitHubRepositoryRef{}
					Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: refName}, o)).To(Succeed())

					result, err := reconcile.NewGetControllerAction(optional, &apiv1.GitHubRepository{}).Execute(ctx, k, o)
					Expect(err).To(MatchError("failed to get owner: Internal error occurred: EOF"))
					Expect(result).ToNot(BeNil())
					Expect(result.Requeue).To(BeFalse())
					Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

					o = &apiv1.GitHubRepositoryRef{}
					Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: refName}, o)).To(Succeed())
					Expect(o.Status.GetInvalidCondition()).To(BeTrueDueTo(apiv1.InternalError))
				},
				Entry("when controller is required", true),
				Entry("when controller is optional", false),
			)
		})
	})
	When("controller reference is not defined", func() {
		var namespace, nonControllerName, refName string
		BeforeEach(func(ctx context.Context) {
			namespace = "default"
			nonControllerName = strings.RandomHash(7)
			refName = strings.RandomHash(7)
			nonController := &apiv1.GitHubRepository{ObjectMeta: metav1.ObjectMeta{Name: nonControllerName, Namespace: namespace, UID: "2"}}
			o := &apiv1.GitHubRepositoryRef{
				ObjectMeta: metav1.ObjectMeta{
					Name:      refName,
					Namespace: namespace,
					OwnerReferences: []metav1.OwnerReference{
						*k8s.NewOwnerReference(nonController, apiv1.GitHubRepositoryGVK, &[]bool{false}[0]),
					},
				},
			}
			k = fake.NewClientBuilder().
				WithScheme(scheme).
				WithIndex(&apiv1.GitHubRepositoryRef{}, k8s.OwnershipIndexField, k8s.IndexGetOwnerReferencesOf).
				WithObjects(nonController, o).
				WithStatusSubresource(o).
				Build()
		})
		When("controller is optional", func() {
			It("should stop", func(ctx context.Context) {
				var o = &apiv1.GitHubRepositoryRef{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: refName}, o)).To(Succeed())

				var controller client.Object
				result, err := reconcile.NewGetControllerAction(true, controller).Execute(ctx, k, o)
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())

				o = &apiv1.GitHubRepositoryRef{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: refName}, o)).To(Succeed())
				Expect(o.Status.IsValid()).To(BeTrue())
			})
		})
		When("controller is required", func() {
			It("should set invalid status and stop", func(ctx context.Context) {
				var o = &apiv1.GitHubRepositoryRef{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: refName}, o)).To(Succeed())

				result, err := reconcile.NewGetControllerAction(false, &apiv1.GitHubRepository{}).Execute(ctx, k, o)
				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.Requeue).To(BeFalse())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

				o = &apiv1.GitHubRepositoryRef{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: refName}, o)).To(Succeed())
				Expect(o.Status.GetInvalidCondition()).To(BeTrueDueTo(apiv1.ControllerMissing))
			})
		})
	})
})
