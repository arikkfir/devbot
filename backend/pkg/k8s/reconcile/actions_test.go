package reconcile_test

import (
	"context"
	"fmt"
	v1 "github.com/arikkfir/devbot/backend/internal/util/testing/api/v1"
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
	DescribeTable("handles failure to get the object",
		func(ctx context.Context, expectedResult ctrl.Result, expectedError, objectGETError error) {
			reconciliation := &reconcile.Reconciliation{
				Actions: []reconcile.Action{
					func(context.Context, client.Client, client.Object) (*ctrl.Result, error) {
						Fail("Action should not execute because object is missing")
						panic("unreachable")
					},
				},
			}
			k := fake.NewClientBuilder().
				WithScheme(scheme).
				WithInterceptorFuncs(interceptor.Funcs{
					Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
						return objectGETError
					},
				}).
				Build()
			result, err := reconciliation.Execute(ctx, k, ctrl.Request{NamespacedName: client.ObjectKey{Namespace: "default", Name: "s"}}, &corev1.Secret{})
			Expect(result).To(Equal(expectedResult))
			if expectedError == nil {
				Expect(err).To(BeNil())
			} else {
				Expect(err).To(MatchError(expectedError))
			}
		},
		Entry("should abort if object is missing", ctrl.Result{}, nil, apierrors.NewNotFound(schema.GroupResource{}, "s")),
		Entry("should abort if object GET fails", ctrl.Result{}, apierrors.NewInternalError(io.EOF), apierrors.NewInternalError(io.EOF)),
	)
	DescribeTable("correctly executes actions",
		func(ctx context.Context, actual []any, expected []any, expectedExecutions int) {
			o := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}}
			key := client.ObjectKeyFromObject(o)
			k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(o).Build()

			executions := 0
			reconciliation := &reconcile.Reconciliation{
				Actions: []reconcile.Action{
					func(context.Context, client.Client, client.Object) (r *ctrl.Result, e error) {
						executions++
						if actual[0] != nil {
							r = actual[0].(*ctrl.Result)
						}
						if actual[1] != nil {
							e = actual[1].(error)
						}
						return
					},
					func(context.Context, client.Client, client.Object) (*ctrl.Result, error) {
						executions++
						return reconcile.Continue()
					},
				},
			}
			result, err := reconciliation.Execute(ctx, k, ctrl.Request{NamespacedName: key}, &corev1.Secret{})
			Expect(executions).To(Equal(expectedExecutions))
			Expect(result).To(Equal(expected[0]))
			if expected[1] == nil {
				Expect(err).To(BeNil())
			} else {
				Expect(err).To(MatchError(expected[1]))
			}
		},
		func(actual []any, expected []any, expectedExecutions int) string {
			return fmt.Sprintf(
				"should execute %d actions and return (%+v, %+v) when first action returns (%+v, %+v)",
				expectedExecutions, expected[0], expected[1], actual[0], actual[1],
			)
		},
		Entry(nil, []any{&ctrl.Result{RequeueAfter: time.Second}, nil}, []any{ctrl.Result{RequeueAfter: time.Second}, nil}, 1),
		Entry(nil, []any{nil, io.EOF}, []any{ctrl.Result{}, io.EOF}, 1),
		Entry(nil, []any{&ctrl.Result{RequeueAfter: 2 * time.Second}, io.EOF}, []any{ctrl.Result{RequeueAfter: 2 * time.Second}, io.EOF}, 1),
		Entry(nil, []any{nil, nil}, []any{ctrl.Result{}, nil}, 2),
	)
})

var _ = Describe("ActionExecution", func() {
	var testMeta metav1.ObjectMeta

	BeforeEach(func() { testMeta = metav1.ObjectMeta{Name: "test", Namespace: "default", Generation: 1} })

	It("should clear conditions before execution for k8s.CommonStatusProvider instances", func(ctx context.Context) {
		var action reconcile.Action = func(_ context.Context, _ client.Client, o client.Object) (*ctrl.Result, error) {
			Expect(o.(*v1.ObjectWithCommonConditions).Status.GetConditions()).To(BeEmpty())
			return reconcile.Continue()
		}
		o := &v1.ObjectWithCommonConditions{
			ObjectMeta: testMeta,
			Status: v1.ObjectWithCommonConditionsStatus{
				Conditions: []metav1.Condition{
					{
						Type:               v1.Invalid,
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 1,
						LastTransitionTime: metav1.Now(),
						Reason:             v1.InternalError,
						Message:            "test",
					},
				},
			},
		}
		k := fake.NewClientBuilder().WithObjects(o).WithStatusSubresource(o).WithScheme(scheme).Build()
		Expect(action.Execute(ctx, k, o)).To(BeNil())
		oo := &v1.ObjectWithCommonConditions{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
		Expect(oo.Status.GetConditions()).To(BeEmpty())
	})

	It("should update status when changed for k8s.CommonStatusProvider instances", func(ctx context.Context) {
		var action reconcile.Action = func(_ context.Context, _ client.Client, o client.Object) (*ctrl.Result, error) {
			o.(*v1.ObjectWithCommonConditions).Status.SetInvalidDueToInternalError("test")
			return reconcile.Continue()
		}
		o := &v1.ObjectWithCommonConditions{ObjectMeta: testMeta}
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(o).WithStatusSubresource(o).Build()
		Expect(action.Execute(ctx, k, o)).To(BeNil())
		oo := &v1.ObjectWithCommonConditions{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
		Expect(oo.Status.GetInvalidReason()).To(Equal(v1.InternalError))
		Expect(oo.Status.GetInvalidMessage()).To(Equal("test"))
	})

	It("should not update status when unchanged for k8s.CommonStatusProvider instances", func(ctx context.Context) {
		var action reconcile.Action = func(_ context.Context, _ client.Client, o client.Object) (*ctrl.Result, error) {
			o.(*v1.ObjectWithCommonConditions).Status.SetInvalidDueToInternalError("test")
			return reconcile.Continue()
		}
		transitionTime := metav1.Time{Time: time.Now().Add(-5 * time.Hour)}
		o := &v1.ObjectWithCommonConditions{
			ObjectMeta: testMeta,
			Status: v1.ObjectWithCommonConditionsStatus{
				Conditions: []metav1.Condition{
					{
						Type:               v1.Invalid,
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 1,
						LastTransitionTime: transitionTime,
						Reason:             v1.InternalError,
						Message:            "test",
					},
				},
			},
		}
		statusUpdated := false
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(o).WithStatusSubresource(o).
			WithInterceptorFuncs(interceptor.Funcs{
				SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, oo client.Object, opts ...client.SubResourceUpdateOption) error {
					statusUpdated = subResourceName == "status" && oo.GetNamespace() == o.Namespace && oo.GetName() == o.Name || statusUpdated
					return c.SubResource(subResourceName).Update(ctx, oo, opts...)
				},
			}).
			Build()
		Expect(action.Execute(ctx, k, o)).To(BeNil())
		Expect(statusUpdated).To(BeFalse())
		oo := &v1.ObjectWithCommonConditions{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
		Expect(oo.Status.GetInvalidReason()).To(Equal(v1.InternalError))
		Expect(oo.Status.GetInvalidMessage()).To(Equal("test"))
		Expect(oo.Status.GetInvalidCondition().LastTransitionTime.Time).To(BeTemporally("~", transitionTime.Time, time.Second))
	})

	It("should allow execution on objects not implementing the k8s.CommonStatusProvider", func(ctx context.Context) {
		var action reconcile.Action = func(context.Context, client.Client, client.Object) (*ctrl.Result, error) { return reconcile.Continue() }
		k := fake.NewClientBuilder().WithScheme(scheme).Build()
		Expect(action.Execute(ctx, k, &v1.ObjectWithoutCommonConditions{ObjectMeta: testMeta})).To(BeNil())
	})

	It("should not update status for objects not implementing k8s.CommonStatusProvider", func(ctx context.Context) {
		var action reconcile.Action = func(context.Context, client.Client, client.Object) (*ctrl.Result, error) { return reconcile.Continue() }
		transitionTime := metav1.Time{Time: time.Now().Add(-5 * time.Hour)}
		o := &v1.ObjectWithoutCommonConditions{
			ObjectMeta: testMeta,
			Status: v1.ObjectWithoutCommonConditionsStatus{
				Conditions: []metav1.Condition{
					{
						Type:               v1.Invalid,
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 1,
						LastTransitionTime: transitionTime,
						Reason:             v1.InternalError,
						Message:            "test",
					},
				},
			},
		}
		statusUpdated := false
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(o).WithStatusSubresource(o).
			WithInterceptorFuncs(interceptor.Funcs{
				SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, oo client.Object, opts ...client.SubResourceUpdateOption) error {
					statusUpdated = subResourceName == "status" && oo.GetNamespace() == o.Namespace && oo.GetName() == o.Name || statusUpdated
					return c.SubResource(subResourceName).Update(ctx, oo, opts...)
				},
			}).
			Build()
		Expect(action.Execute(ctx, k, o)).To(BeNil())
		Expect(statusUpdated).To(BeFalse())
		oo := &v1.ObjectWithoutCommonConditions{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
		Expect(oo.Status.GetInvalidReason()).To(Equal(v1.InternalError))
		Expect(oo.Status.GetInvalidMessage()).To(Equal("test"))
		Expect(oo.Status.GetInvalidCondition().LastTransitionTime.Time).To(BeTemporally("~", transitionTime.Time, time.Second))
	})

	It("should abort processing when status update fails due to object not found", func(ctx context.Context) {
		var action reconcile.Action = func(_ context.Context, _ client.Client, o client.Object) (*ctrl.Result, error) {
			o.(*v1.ObjectWithCommonConditions).Status.SetInvalidDueToInternalError("test")
			return reconcile.Continue()
		}
		o := &v1.ObjectWithCommonConditions{ObjectMeta: testMeta}
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(o).WithStatusSubresource(o).
			WithInterceptorFuncs(interceptor.Funcs{
				SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, oo client.Object, opts ...client.SubResourceUpdateOption) error {
					if subResourceName == "status" && oo.GetNamespace() == o.Namespace && oo.GetName() == o.Name {
						return apierrors.NewNotFound(schema.GroupResource{}, oo.GetName())
					}
					return c.SubResource(subResourceName).Update(ctx, oo, opts...)
				},
			}).
			Build()
		Expect(action.Execute(ctx, k, o)).To(Equal(&ctrl.Result{}))
	})

	It("should requeue when status update fails due to conflict", func(ctx context.Context) {
		var action reconcile.Action = func(_ context.Context, _ client.Client, o client.Object) (*ctrl.Result, error) {
			o.(*v1.ObjectWithCommonConditions).Status.SetInvalidDueToInternalError("test")
			return reconcile.Continue()
		}
		o := &v1.ObjectWithCommonConditions{ObjectMeta: testMeta}
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(o).WithStatusSubresource(o).
			WithInterceptorFuncs(interceptor.Funcs{
				SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, oo client.Object, opts ...client.SubResourceUpdateOption) error {
					if subResourceName == "status" && oo.GetNamespace() == o.Namespace && oo.GetName() == o.Name {
						return apierrors.NewConflict(schema.GroupResource{}, oo.GetName(), io.EOF)
					}
					return c.SubResource(subResourceName).Update(ctx, oo, opts...)
				},
			}).
			Build()
		Expect(action.Execute(ctx, k, o)).To(Equal(&ctrl.Result{Requeue: true}))
	})

	It("should requeue when status update fails due to an unexpected error", func(ctx context.Context) {
		var action reconcile.Action = func(_ context.Context, _ client.Client, o client.Object) (*ctrl.Result, error) {
			o.(*v1.ObjectWithCommonConditions).Status.SetInvalidDueToInternalError("test")
			return reconcile.Continue()
		}
		o := &v1.ObjectWithCommonConditions{ObjectMeta: testMeta}
		k := fake.NewClientBuilder().WithScheme(scheme).WithObjects(o).WithStatusSubresource(o).
			WithInterceptorFuncs(interceptor.Funcs{
				SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, oo client.Object, opts ...client.SubResourceUpdateOption) error {
					if subResourceName == "status" && oo.GetNamespace() == o.Namespace && oo.GetName() == o.Name {
						return apierrors.NewInternalError(io.EOF)
					}
					return c.SubResource(subResourceName).Update(ctx, oo, opts...)
				},
			}).
			Build()
		result, err := action.Execute(ctx, k, o)
		Expect(result).To(Equal(&ctrl.Result{}))
		Expect(err).To(MatchError("failed to update object status: Internal error occurred: EOF"))
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
