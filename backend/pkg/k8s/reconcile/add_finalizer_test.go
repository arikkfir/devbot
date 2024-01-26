package reconcile_test

import (
	"context"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	v1 "github.com/arikkfir/devbot/backend/internal/util/testing/api/v1"
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

var _ = Describe("NewAddFinalizerAction", func() {
	var k client.Client
	var namespace, name, finalizer string
	BeforeEach(func() {
		namespace = "default"
		name = strings.RandomHash(7)
		finalizer = strings.RandomHash(7)
	})
	When("finalizer is not present", func() {
		When("update succeeds", func() {
			BeforeEach(func() {
				k = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(&v1.ObjectWithCommonConditions{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}).
					Build()
			})
			It("should add the finalizer", func(ctx context.Context) {
				o := &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				result, err := reconcile.NewAddFinalizerAction(finalizer).Execute(ctx, k, o)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Requeue).To(BeTrue())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

				o2 := &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o2)).To(Succeed())
				Expect(o2.Finalizers).To(ConsistOf(finalizer))
			})
		})
		When("update fails because object is not found", func() {
			BeforeEach(func() {
				k = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(&v1.ObjectWithCommonConditions{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}).
					WithInterceptorFuncs(interceptor.Funcs{
						Update: func(ctx context.Context, c client.WithWatch, o client.Object, opts ...client.UpdateOption) error {
							return apierrors.NewNotFound(schema.GroupResource{
								Group:    o.GetObjectKind().GroupVersionKind().Group,
								Resource: o.GetObjectKind().GroupVersionKind().Kind,
							}, name)
						},
					}).
					Build()
			})
			It("should abort processing", func(ctx context.Context) {
				o := &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				result, err := reconcile.NewAddFinalizerAction(finalizer).Execute(ctx, k, o)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

				o2 := &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o2)).To(Succeed())
				Expect(o2.Finalizers).To(BeEmpty())
			})
		})
		When("update fails because object was modified externally", func() {
			BeforeEach(func() {
				k = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(&v1.ObjectWithCommonConditions{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}).
					WithInterceptorFuncs(interceptor.Funcs{
						Update: func(ctx context.Context, c client.WithWatch, o client.Object, opts ...client.UpdateOption) error {
							return apierrors.NewConflict(schema.GroupResource{
								Group:    o.GetObjectKind().GroupVersionKind().Group,
								Resource: o.GetObjectKind().GroupVersionKind().Kind,
							}, name, io.EOF)
						},
					}).
					Build()
			})
			It("should requeue immediately", func(ctx context.Context) {
				o := &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				result, err := reconcile.NewAddFinalizerAction(finalizer).Execute(ctx, k, o)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Requeue).To(BeTrue())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

				o2 := &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o2)).To(Succeed())
				Expect(o2.Finalizers).To(BeEmpty())
			})
		})
		When("update fails because of an unknown error", func() {
			BeforeEach(func() {
				o := &v1.ObjectWithCommonConditions{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
				k = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(o).
					WithInterceptorFuncs(interceptor.Funcs{
						Update: func(ctx context.Context, c client.WithWatch, o client.Object, opts ...client.UpdateOption) error {
							return apierrors.NewInternalError(io.EOF)
						},
					}).
					WithStatusSubresource(o).
					Build()
			})
			It("should set invalid condition and requeue immediately", func(ctx context.Context) {
				o := &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				result, err := reconcile.NewAddFinalizerAction(finalizer).Execute(ctx, k, o)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Requeue).To(BeTrue())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

				oo := &v1.ObjectWithCommonConditions{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, oo)).To(Succeed())
				Expect(oo.Finalizers).To(BeEmpty())
				Expect(oo.Status.GetInvalidCondition()).To(BeTrueDueTo(v1.AddFinalizerFailed))
			})
		})
	})
	When("finalizer is already present", func() {
		BeforeEach(func() {
			k = fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(&v1.ObjectWithCommonConditions{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{finalizer},
						Name:       name,
						Namespace:  namespace,
					},
				}).
				Build()
		})
		It("should keep finalizer", func(ctx context.Context) {
			o := &v1.ObjectWithCommonConditions{}
			Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
			result, err := reconcile.NewAddFinalizerAction(finalizer).Execute(ctx, k, o)
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())

			oo := &v1.ObjectWithCommonConditions{}
			Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, oo)).To(Succeed())
			Expect(oo.Finalizers).To(ConsistOf(finalizer))
		})
	})
})
