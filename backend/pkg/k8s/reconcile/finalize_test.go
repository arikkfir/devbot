package reconcile_test

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/secureworks/errors"
	"io"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"time"
)

var _ = Describe("NewFinalizeAction", func() {
	var k client.WithWatch
	var namespace, name, finalizer string
	var finalizeFunc func() error

	BeforeEach(func() {
		namespace = "default"
		name = strings.RandomHash(7)
		finalizer = strings.RandomHash(7)
	})

	const deletionPreventionFinalizer = "finalizer-to-prevent-actual-deletion"
	validateObjectUpdate := func() {
		When("update succeeds", func() {
			It("should remove finalizer and stop", func(ctx context.Context) {
				o := &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				result, err := reconcile.NewFinalizeAction(finalizer, finalizeFunc).Execute(ctx, k, o)
				Expect(err).To(BeNil())
				Expect(result.Requeue).To(BeFalse())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

				o = &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				Expect(o.Finalizers).To(ConsistOf(deletionPreventionFinalizer))
				Expect(o.Status.GetInvalidCondition()).To(BeNil())
			})
		})
		When("update fails due to object not found", func() {
			BeforeEach(func() {
				k = interceptor.NewClient(k, interceptor.Funcs{
					Update: func(ctx context.Context, c client.WithWatch, o client.Object, opts ...client.UpdateOption) error {
						return apierrors.NewNotFound(schema.GroupResource{
							Group:    o.GetObjectKind().GroupVersionKind().Group,
							Resource: o.GetObjectKind().GroupVersionKind().Kind,
						}, name)
					},
				})
			})
			It("should stop", func(ctx context.Context) {
				o := &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				result, err := reconcile.NewFinalizeAction(finalizer, finalizeFunc).Execute(ctx, k, o)
				Expect(err).To(BeNil())
				Expect(result.Requeue).To(BeFalse())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))
			})
		})
		When("update fails due to conflict", func() {
			BeforeEach(func() {
				k = interceptor.NewClient(k, interceptor.Funcs{
					Update: func(ctx context.Context, c client.WithWatch, o client.Object, opts ...client.UpdateOption) error {
						return apierrors.NewConflict(schema.GroupResource{
							Group:    o.GetObjectKind().GroupVersionKind().Group,
							Resource: o.GetObjectKind().GroupVersionKind().Kind,
						}, name, io.EOF)
					},
				})
			})
			It("should requeue", func(ctx context.Context) {
				o := &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				result, err := reconcile.NewFinalizeAction(finalizer, finalizeFunc).Execute(ctx, k, o)
				Expect(err).To(BeNil())
				Expect(result.Requeue).To(BeTrue())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))
			})
		})
		When("update fails due to unknown error", func() {
			BeforeEach(func() {
				k = interceptor.NewClient(k, interceptor.Funcs{
					Update: func(ctx context.Context, c client.WithWatch, o client.Object, opts ...client.UpdateOption) error {
						return apierrors.NewInternalError(io.EOF)
					},
				})
			})
			It("should set invalid condition and requeue", func(ctx context.Context) {
				o := &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				result, err := reconcile.NewFinalizeAction(finalizer, finalizeFunc).Execute(ctx, k, o)
				Expect(err).To(BeNil())
				Expect(result.Requeue).To(BeTrue())
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

				o = &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				Expect(o.Finalizers).To(ConsistOf(finalizer, deletionPreventionFinalizer))
				Expect(o.Status.GetInvalidCondition()).To(BeTrueDueTo(apiv1.InternalError))
			})
		})
	}

	When("object is deleted", func() {
		When("finalizer is present", func() {
			BeforeEach(func() {
				src := &apiv1.GitHubRepository{
					ObjectMeta: metav1.ObjectMeta{
						DeletionTimestamp: &[]metav1.Time{metav1.Now()}[0],
						Finalizers:        []string{finalizer, deletionPreventionFinalizer},
						Name:              name,
						Namespace:         namespace,
					}}
				k = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(src).
					WithStatusSubresource(src).
					Build()
			})
			When("finalize function is given", func() {
				When("finalize function fails", func() {
					BeforeEach(func() { finalizeFunc = func() error { return errors.New("finalization function failure") } })
					It("should set invalid condition and requeue", func(ctx context.Context) {
						o := &apiv1.GitHubRepository{}
						Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
						result, err := reconcile.NewFinalizeAction(finalizer, finalizeFunc).Execute(ctx, k, o)
						Expect(err).To(BeNil())
						Expect(result.Requeue).To(BeTrue())
						Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

						o = &apiv1.GitHubRepository{}
						Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
						Expect(o.Status.GetInvalidCondition()).To(BeTrueDueTo(apiv1.FinalizationFailed))
					})
				})
				When("finalize function succeeds", func() {
					BeforeEach(func() { finalizeFunc = func() error { return nil } })
					validateObjectUpdate()
				})
			})
			When("finalize function is not given", func() {
				BeforeEach(func() { finalizeFunc = nil })
				validateObjectUpdate()
			})
		})
		When("finalizer is not present", func() {
			BeforeEach(func() {
				src := &apiv1.GitHubRepository{
					ObjectMeta: metav1.ObjectMeta{
						DeletionTimestamp: &[]metav1.Time{metav1.Now()}[0],
						Finalizers:        []string{deletionPreventionFinalizer},
						Name:              name,
						Namespace:         namespace,
					}}
				k = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(src).
					WithStatusSubresource(src).
					Build()
				finalizeFunc = nil
			})
			It("should just stop", func(ctx context.Context) {
				o := &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
				result, err := reconcile.NewFinalizeAction(finalizer, finalizeFunc).Execute(ctx, k, o)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(&ctrl.Result{}))
			})
		})
	})
	When("object is not deleted", func() {
		BeforeEach(func() {
			src := &apiv1.GitHubRepository{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
			k = fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(src).
				WithStatusSubresource(src).
				Build()
			finalizeFunc = nil
		})
		It("should continue", func(ctx context.Context) {
			o := &apiv1.GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, o)).To(Succeed())
			result, err := reconcile.NewFinalizeAction(finalizer, finalizeFunc).Execute(ctx, k, o)
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
		})
	})
})
