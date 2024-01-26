package reconcile_test

import (
	"context"
	v1 "github.com/arikkfir/devbot/backend/internal/util/testing/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("NewGetOwnedObjectsOfTypeAction", func() {
	var k client.WithWatch
	BeforeEach(func(ctx context.Context) {
		owner1 := &v1.ObjectWithCommonConditions{ObjectMeta: metav1.ObjectMeta{Finalizers: []string{"test"}, Name: "owner1", Namespace: "default", UID: "1"}}
		owner2 := &v1.ObjectWithCommonConditions{ObjectMeta: metav1.ObjectMeta{Finalizers: []string{"test"}, Name: "owner2", Namespace: "default", UID: "2"}}
		child1 := &v1.ObjectWithCommonConditions{ObjectMeta: metav1.ObjectMeta{Finalizers: []string{"test"}, Name: "child1", Namespace: "default"}}
		Expect(controllerutil.SetOwnerReference(owner1, child1, scheme)).To(Succeed())
		child2 := &v1.ObjectWithCommonConditions{ObjectMeta: metav1.ObjectMeta{Finalizers: []string{"test"}, Name: "child2", Namespace: "default"}}
		Expect(controllerutil.SetOwnerReference(owner2, child2, scheme)).To(Succeed())
		k = fake.NewClientBuilder().
			WithScheme(scheme).
			WithIndex(&v1.ObjectWithCommonConditions{}, k8s.OwnershipIndexField, k8s.IndexGetOwnerReferencesOf).
			WithObjects(owner1, owner2, child1, child2).
			WithStatusSubresource(owner1, owner2).
			Build()
	})
	When("owner and owned are defined", func() {
		When("listing owned objects fails", func() {
			BeforeEach(func(ctx context.Context) {
				k = interceptor.NewClient(k, interceptor.Funcs{
					List: func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
						return apierrors.NewInternalError(io.EOF)
					},
				})
			})
			It("should fail with an error", func(ctx context.Context) {
				var owner = &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: "default", Name: "owner1"}, owner)).To(Succeed())
				var child1 = &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Namespace: "default", Name: "child1"}, child1)).To(Succeed())

				var owned = &v1.ObjectWithCommonConditionsList{}
				result, err := reconcile.NewGetOwnedObjectsOfTypeAction(owned).Execute(ctx, k, owner)
				Expect(err).ToNot(BeNil())
				Expect(result).To(Equal(&ctrl.Result{}))
			})
		})
		It("should find owned objects", func(ctx context.Context) {
			var owner = &v1.ObjectWithCommonConditions{}
			Expect(k.Get(ctx, client.ObjectKey{Namespace: "default", Name: "owner1"}, owner)).To(Succeed())
			var child1 = &v1.ObjectWithCommonConditions{}
			Expect(k.Get(ctx, client.ObjectKey{Namespace: "default", Name: "child1"}, child1)).To(Succeed())

			var owned = &v1.ObjectWithCommonConditionsList{}
			result, err := reconcile.NewGetOwnedObjectsOfTypeAction(owned).Execute(ctx, k, owner)
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
			Expect(owned.Items).To(HaveLen(1))
			Expect(owned.Items[0].Name).To(Equal(child1.Name))
			Expect(owned.Items[0].Namespace).To(Equal(child1.Namespace))
		})
	})
})
