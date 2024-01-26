package reconcile_test

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	corev1 "k8s.io/api/core/v1"
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
					WithObjects(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}).
					Build()
			})
			It("should add the finalizer", func(ctx context.Context) {
				o := &corev1.ConfigMap{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				result, err := reconcile.NewAddFinalizerAction(finalizer).Execute(ctx, k, o)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Requeue).To(BeTrue())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

				o2 := &corev1.ConfigMap{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o2)).To(Succeed())
				Expect(o2.Finalizers).To(ConsistOf(finalizer))
			})
		})
		When("update fails because object is not found", func() {
			BeforeEach(func() {
				k = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}).
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
				o := &corev1.ConfigMap{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				result, err := reconcile.NewAddFinalizerAction(finalizer).Execute(ctx, k, o)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

				o2 := &corev1.ConfigMap{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o2)).To(Succeed())
				Expect(o2.Finalizers).To(BeEmpty())
			})
		})
		When("update fails because object was modified externally", func() {
			BeforeEach(func() {
				k = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}).
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
				o := &corev1.ConfigMap{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				result, err := reconcile.NewAddFinalizerAction(finalizer).Execute(ctx, k, o)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Requeue).To(BeTrue())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

				o2 := &corev1.ConfigMap{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o2)).To(Succeed())
				Expect(o2.Finalizers).To(BeEmpty())
			})
		})
		When("update fails because of an unknown error", func() {
			BeforeEach(func() {
				o := &apiv1.GitHubRepository{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
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
				o := &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				result, err := reconcile.NewAddFinalizerAction(finalizer).Execute(ctx, k, o)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Requeue).To(BeTrue())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

				o2 := &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o2)).To(Succeed())
				Expect(o2.Finalizers).To(BeEmpty())
				Expect(o2.Status.GetInvalidCondition()).To(BeTrueDueTo(apiv1.InternalError))
			})
		})
	})
	When("finalizer is already present", func() {
		BeforeEach(func() {
			k = fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{finalizer},
						Name:       name,
						Namespace:  namespace,
					},
				}).
				Build()
		})
		It("should keep finalizer", func(ctx context.Context) {
			o := &corev1.ConfigMap{}
			Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
			result, err := reconcile.NewAddFinalizerAction(finalizer).Execute(ctx, k, o)
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())

			o2 := &corev1.ConfigMap{}
			Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o2)).To(Succeed())
			Expect(o2.Finalizers).To(ConsistOf(finalizer))
		})
	})
})
