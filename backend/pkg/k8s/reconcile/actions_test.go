package reconcile_test

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"time"
)

var _ = Describe("Reconciliation", func() {
	When("object not found", func() {
		var s *corev1.Secret
		var k client.WithWatch
		BeforeEach(func(ctx context.Context) {
			s = &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}}
			k = fake.NewClientBuilder().
				WithScheme(scheme).
				WithInterceptorFuncs(interceptor.Funcs{
					Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						return apierrors.NewNotFound(schema.GroupResource{}, s.Name)
					},
				}).
				Build()
		})
		It("should abort", func(ctx context.Context) {
			executed := false
			reconciliation := &reconcile.Reconciliation{
				Actions: []reconcile.Action{
					func(context.Context, client.Client, client.Object) (*ctrl.Result, error) {
						executed = true
						return reconcile.Continue()
					},
				},
			}

			result, err := reconciliation.Execute(ctx, k, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(s)}, &corev1.Secret{})
			Expect(executed).To(BeFalse())
			Expect(err).To(BeNil())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})
	When("object exists", func() {
		var s *corev1.Secret
		var k client.Client
		BeforeEach(func(ctx context.Context) {
			s = &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(s).Build()
		})
		When("an action fails", func() {
			It("should return its result and error", func(ctx context.Context) {
				reconciliation := &reconcile.Reconciliation{
					Actions: []reconcile.Action{
						func(context.Context, client.Client, client.Object) (*ctrl.Result, error) {
							return reconcile.RequeueDueToError(io.EOF)
						},
						func(context.Context, client.Client, client.Object) (*ctrl.Result, error) {
							Fail("2nd action should not be executed")
							panic("unreachable")
						},
					},
				}
				ss := &corev1.Secret{}
				Expect(k.Get(ctx, client.ObjectKeyFromObject(s), ss)).To(Succeed())
				result, err := reconciliation.Execute(ctx, k, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(ss)}, ss)
				Expect(err).To(MatchError(io.EOF))
				Expect(result).To(Equal(ctrl.Result{}))
			})
		})
		When("an action succeeds with a result", func() {
			It("should return its result", func(ctx context.Context) {
				reconciliation := &reconcile.Reconciliation{
					Actions: []reconcile.Action{
						func(context.Context, client.Client, client.Object) (*ctrl.Result, error) {
							return reconcile.RequeueAfter(5 * time.Second)
						},
						func(context.Context, client.Client, client.Object) (*ctrl.Result, error) {
							Fail("2nd action should not be executed")
							panic("unreachable")
						},
					},
				}
				ss := &corev1.Secret{}
				Expect(k.Get(ctx, client.ObjectKeyFromObject(s), ss)).To(Succeed())
				result, err := reconciliation.Execute(ctx, k, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(ss)}, ss)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(ctrl.Result{RequeueAfter: 5 * time.Second}))
			})
		})
		When("all actions succeed", func() {
			It("should return an empty result and no error", func(ctx context.Context) {
				executions := 0
				reconciliation := &reconcile.Reconciliation{
					Actions: []reconcile.Action{
						func(context.Context, client.Client, client.Object) (*ctrl.Result, error) {
							executions++
							return reconcile.Continue()
						},
						func(context.Context, client.Client, client.Object) (*ctrl.Result, error) {
							executions++
							return reconcile.Continue()
						},
					},
				}
				ss := &corev1.Secret{}
				Expect(k.Get(ctx, client.ObjectKeyFromObject(s), ss)).To(Succeed())
				result, err := reconciliation.Execute(ctx, k, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(ss)}, ss)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(executions).To(Equal(2))
			})
		})
	})
})

var _ = Describe("Action", func() {
	var k client.WithWatch
	var namespace, repoName string
	BeforeEach(func(ctx context.Context) {
		namespace = "default"
		repoName = "test"
		o := &apiv1.GitHubRepository{ObjectMeta: metav1.ObjectMeta{Name: repoName, Namespace: namespace}}
		k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(o).WithStatusSubresource(o).Build()
	})
	When("an object has a pre-existing condition", func() {
		BeforeEach(func(ctx context.Context) {
			o := &apiv1.GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKey{Name: repoName, Namespace: namespace}, o)).To(Succeed())
			o.Status.SetInvalidDueToInternalError("test")
			Expect(k.Status().Update(ctx, o)).To(Succeed())
		})
		It("should clear the pre-existing conditions", func(ctx context.Context) {
			var action reconcile.Action
			action = func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
				return reconcile.Continue()
			}

			o := &apiv1.GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKey{Name: repoName, Namespace: namespace}, o)).To(Succeed())
			Expect(o.Status.IsInvalid()).To(BeTrue()) // update succeeded

			result, err := action.Execute(ctx, k, o)
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())

			o = &apiv1.GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKey{Name: repoName, Namespace: namespace}, o)).To(Succeed())
			Expect(o.Status.GetInvalidCondition()).To(BeNil())
		})
	})
	When("action sets a condition", func() {
		var action reconcile.Action
		BeforeEach(func(ctx context.Context) {
			action = func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
				o.(*apiv1.GitHubRepository).Status.SetInvalidDueToInternalError("test")
				return reconcile.Continue()
			}
		})
		When("status update succeeds", func() {
			It("should update the object status", func(ctx context.Context) {
				o := &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Name: repoName, Namespace: namespace}, o)).To(Succeed())

				result, err := action.Execute(ctx, k, o)
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())

				oo := &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
				Expect(oo.Status.GetInvalidCondition()).To(BeTrueDueTo(apiv1.InternalError, "test"))
			})
		})
		When("status update fails due to object missing", func() {
			It("should ignore the error", func(ctx context.Context) {
				k := interceptor.NewClient(k, interceptor.Funcs{
					SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, o client.Object, opts ...client.SubResourceUpdateOption) error {
						if subResourceName == "status" {
							return apierrors.NewNotFound(schema.GroupResource{
								Group:    o.GetObjectKind().GroupVersionKind().Group,
								Resource: o.GetObjectKind().GroupVersionKind().Kind,
							}, o.GetName())
						}
						return c.SubResource(subResourceName).Update(ctx, o, opts...)
					},
				})

				o := &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Name: repoName, Namespace: namespace}, o)).To(Succeed())
				result, err := action.Execute(ctx, k, o)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(&ctrl.Result{}))
			})
		})
		When("status update fails due to conflict", func() {
			It("should requeue", func(ctx context.Context) {
				k := interceptor.NewClient(k, interceptor.Funcs{
					SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, o client.Object, opts ...client.SubResourceUpdateOption) error {
						if subResourceName == "status" {
							return apierrors.NewConflict(schema.GroupResource{
								Group:    o.GetObjectKind().GroupVersionKind().Group,
								Resource: o.GetObjectKind().GroupVersionKind().Kind,
							}, o.GetName(), io.EOF)
						}
						return c.SubResource(subResourceName).Update(ctx, o, opts...)
					},
				})
				o := &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Name: repoName, Namespace: namespace}, o)).To(Succeed())
				result, err := action.Execute(ctx, k, o)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(&ctrl.Result{Requeue: true}))
			})
		})
		When("status update fails due to unknown error", func() {
			BeforeEach(func(ctx context.Context) {
				k = interceptor.NewClient(k, interceptor.Funcs{
					SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, o client.Object, opts ...client.SubResourceUpdateOption) error {
						if subResourceName == "status" {
							return apierrors.NewInternalError(io.EOF)
						}
						return c.SubResource(subResourceName).Update(ctx, o, opts...)
					},
				})
			})
			It("should requeue with error", func(ctx context.Context) {
				o := &apiv1.GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKey{Name: repoName, Namespace: namespace}, o)).To(Succeed())
				result, err := action.Execute(ctx, k, o)
				Expect(err).ToNot(BeNil())
				Expect(result).To(Equal(&ctrl.Result{}))
			})
		})
	})
})

var _ = Describe("Continue", func() {
	It("should return no result and no error", func() {
		result, err := reconcile.Continue()
		Expect(err).To(BeNil())
		Expect(result).To(BeNil())
	})
})

var _ = Describe("DoNotRequeue", func() {
	It("should return an empty result and no error", func() {
		result, err := reconcile.DoNotRequeue()
		Expect(err).To(BeNil())
		Expect(result).To(Equal(&ctrl.Result{}))
	})
})

var _ = Describe("Requeue", func() {
	It("should request a requeue with no error", func() {
		result, err := reconcile.Requeue()
		Expect(err).To(BeNil())
		Expect(result).To(Equal(&ctrl.Result{Requeue: true}))
	})
})

var _ = Describe("RequeueAfter", func() {
	It("should request a requeue after the given duration, with no error", func() {
		result, err := reconcile.RequeueAfter(10)
		Expect(err).To(BeNil())
		Expect(result).To(Equal(&ctrl.Result{RequeueAfter: 10}))
	})
})

var _ = Describe("RequeueDueToError", func() {
	It("should return an empty result with the given error", func() {
		result, err := reconcile.RequeueDueToError(io.EOF)
		Expect(err).To(MatchError(io.EOF))
		Expect(result).To(Equal(&ctrl.Result{}))
	})
})
