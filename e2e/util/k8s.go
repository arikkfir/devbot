package util

import (
	"context"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/arikkfir/devbot/api/v1"
	"github.com/arikkfir/devbot/internal/util/strings"
)

var (
	DevbotNamespace                    = "devbot"
	DevbotControllerServiceAccountName = "devbot-controller"
	k8sServiceAccountKind              = reflect.TypeOf(corev1.ServiceAccount{}).Name()
	k8sClusterRoleKind                 = reflect.TypeOf(rbacv1.ClusterRole{}).Name()
)

func CreateK8sNamespace(ctx context.Context, c client.Client) string {
	GinkgoHelper()
	ns := &corev1.Namespace{ObjectMeta: ctrl.ObjectMeta{Name: strings.RandomHash(7), Labels: map[string]string{"devbot.kfirs.com/purpose": "test"}}}
	Expect(c.Create(ctx, ns)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(c.Delete(ctx, ns)).To(Succeed()) })
	return ns.Name
}

func CreateK8sApplication(ctx context.Context, c client.Client, ns string, spec apiv1.ApplicationSpec) string {
	GinkgoHelper()
	app := &apiv1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      strings.RandomHash(7),
		},
		Spec: spec,
	}
	Expect(c.Create(ctx, app)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(c.Delete(ctx, app)).To(Succeed()) })
	return app.Name
}
