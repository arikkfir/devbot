package reconcile_test

import (
	"context"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("NewSaveObjectReferenceAction", func() {
	It("should store object in given pointer and continue", func(ctx context.Context) {
		var c client.Object
		o := &corev1.ConfigMap{}
		result, err := reconcile.NewSaveObjectReferenceAction(&c).Execute(ctx, fake.NewFakeClient(), o)
		Expect(err).To(BeNil())
		Expect(result).To(BeNil())
		Expect(c).To(BeIdenticalTo(o))
	})
})
