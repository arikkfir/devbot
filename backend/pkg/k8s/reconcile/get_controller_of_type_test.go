package reconcile_test

import (
	"context"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	v1 "github.com/arikkfir/devbot/backend/internal/util/testing/api/v1"
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
		var namespace, controllerName, nonControllerName, controlleeName string
		BeforeEach(func(ctx context.Context) {
			namespace = "default"
			controllerName = strings.RandomHash(7)
			nonControllerName = strings.RandomHash(7)
			controlleeName = strings.RandomHash(7)
			controller := &v1.ObjectWithCommonConditions{ObjectMeta: metav1.ObjectMeta{Name: controllerName, Namespace: namespace, UID: "1"}}
			nonController := &v1.ObjectWithCommonConditions{ObjectMeta: metav1.ObjectMeta{Name: nonControllerName, Namespace: namespace, UID: "2"}}
			controllee := &v1.ObjectWithCommonConditions{
				ObjectMeta: metav1.ObjectMeta{
					Name:      controlleeName,
					Namespace: namespace,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(controller, v1.ObjectWithCommonConditionsGVK),
						*k8s.NewOwnerReference(nonController, v1.ObjectWithCommonConditionsGVK, &[]bool{false}[0]),
					},
				},
			}
			k = fake.NewClientBuilder().
				WithScheme(scheme).
				WithIndex(&v1.ObjectWithCommonConditions{}, k8s.OwnershipIndexField, k8s.IndexGetOwnerReferencesOf).
				WithObjects(controller, nonController, controllee).
				WithStatusSubresource(controllee).
				Build()
		})
		When("controller exists", func() {
			It("should obtain controller and continue", func(ctx context.Context) {
				var expectedController = &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: controllerName}, expectedController)).To(Succeed())

				controllee := &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: controlleeName}, controllee)).To(Succeed())

				var actualController = &v1.ObjectWithCommonConditions{}
				result, err := reconcile.NewGetControllerAction(false, controllee, actualController).Execute(ctx, k, controllee)
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
					var o = &v1.ObjectWithCommonConditions{}
					Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: controlleeName}, o)).To(Succeed())

					result, err := reconcile.NewGetControllerAction(false, o, &v1.ObjectWithCommonConditions{}).Execute(ctx, k, o)
					Expect(err).To(BeNil())
					Expect(result).ToNot(BeNil())
					Expect(result.Requeue).To(BeFalse())
					Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

					oo := &v1.ObjectWithCommonConditions{}
					Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: controlleeName}, oo)).To(Succeed())
					Expect(oo.Status.GetInvalidCondition()).To(BeTrueDueTo(v1.ControllerMissing))
				})
			})
			When("controller is optional", func() {
				It("should stop", func(ctx context.Context) {
					var o = &v1.ObjectWithCommonConditions{}
					Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: controlleeName}, o)).To(Succeed())

					var controller client.Object
					result, err := reconcile.NewGetControllerAction(true, o, controller).Execute(ctx, k, o)
					Expect(err).To(BeNil())
					Expect(result).To(BeNil())

					oo := &v1.ObjectWithCommonConditions{}
					Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: controlleeName}, oo)).To(Succeed())
					Expect(oo.Status.IsValid()).To(BeTrue())
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
					var o = &v1.ObjectWithCommonConditions{}
					Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: controlleeName}, o)).To(Succeed())

					result, err := reconcile.NewGetControllerAction(optional, o, &v1.ObjectWithCommonConditions{}).Execute(ctx, k, o)
					Expect(err).To(MatchError("failed to get owner: Internal error occurred: EOF"))
					Expect(result).ToNot(BeNil())
					Expect(result.Requeue).To(BeFalse())
					Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

					oo := &v1.ObjectWithCommonConditions{}
					Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: controlleeName}, oo)).To(Succeed())
					Expect(oo.Status.GetInvalidCondition()).To(BeTrueDueTo(v1.InternalError))
				},
				Entry("when controller is required", true),
				Entry("when controller is optional", false),
			)
		})
	})
	When("controller reference is not defined", func() {
		var namespace, nonControllerName, controlleeName string
		BeforeEach(func(ctx context.Context) {
			namespace = "default"
			nonControllerName = strings.RandomHash(7)
			controlleeName = strings.RandomHash(7)
			nonController := &v1.ObjectWithCommonConditions{ObjectMeta: metav1.ObjectMeta{Name: nonControllerName, Namespace: namespace, UID: "2"}}
			o := &v1.ObjectWithCommonConditions{
				ObjectMeta: metav1.ObjectMeta{
					Name:      controlleeName,
					Namespace: namespace,
					OwnerReferences: []metav1.OwnerReference{
						*k8s.NewOwnerReference(nonController, v1.ObjectWithCommonConditionsGVK, &[]bool{false}[0]),
					},
				},
			}
			k = fake.NewClientBuilder().
				WithScheme(scheme).
				WithIndex(&v1.ObjectWithCommonConditions{}, k8s.OwnershipIndexField, k8s.IndexGetOwnerReferencesOf).
				WithObjects(nonController, o).
				WithStatusSubresource(o).
				Build()
		})
		When("controller is optional", func() {
			It("should stop", func(ctx context.Context) {
				var o = &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: controlleeName}, o)).To(Succeed())

				var controller client.Object
				result, err := reconcile.NewGetControllerAction(true, o, controller).Execute(ctx, k, o)
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())

				oo := &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: controlleeName}, oo)).To(Succeed())
				Expect(oo.Status.IsValid()).To(BeTrue())
			})
		})
		When("controller is required", func() {
			It("should set invalid status and stop", func(ctx context.Context) {
				var o = &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: controlleeName}, o)).To(Succeed())

				result, err := reconcile.NewGetControllerAction(false, o, &v1.ObjectWithCommonConditions{}).Execute(ctx, k, o)
				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.Requeue).To(BeFalse())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

				oo := &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: controlleeName}, oo)).To(Succeed())
				Expect(oo.Status.GetInvalidCondition()).To(BeTrueDueTo(v1.ControllerMissing))
			})
		})
	})
})
