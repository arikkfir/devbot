package reconcile_test

import (
	"context"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"time"
)

var _ = Describe("NewRequeueAfterAction", func() {
	It("should requeue with given refresh interval", func(ctx context.Context) {
		result, err := reconcile.NewRequeueAfterAction(time.Minute).Execute(ctx, fake.NewFakeClient(), &corev1.ConfigMap{})
		Expect(err).To(BeNil())
		Expect(result).To(Equal(&ctrl.Result{RequeueAfter: time.Minute}))
	})
})
