package util

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	apiv1 "github.com/arikkfir/devbot/api/v1"
	"github.com/arikkfir/devbot/internal/util/k8s"
)

func NewK8sClient() (client.Client, *rest.Config) {
	GinkgoHelper()
	ctx, cancel := context.WithCancel(context.Background())
	DeferCleanup(cancel)

	userHomeDir, err := os.UserHomeDir()
	Expect(err).To(BeNil())

	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	Expect(err).To(BeNil())

	scheme := runtime.NewScheme()
	utilruntime.Must(apiv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Logger: GinkgoLogr,
		Scheme: scheme,
		Client: client.Options{
			Scheme: scheme,
			Cache:  &client.CacheOptions{
				// DisableFor: []client.Object{&corev1.Secret{}},
			},
		},
		LeaderElection:         false,
		Metrics:                metricsserver.Options{BindAddress: "0"},
		HealthProbeBindAddress: "0",
		PprofBindAddress:       "0",
	})
	Expect(err).To(BeNil())

	nameIndexFunc := func(o client.Object) []string { return []string{o.GetName()} }
	preferredBranchIndexFunc := func(o client.Object) []string { return []string{o.(*apiv1.Environment).Spec.PreferredBranch} }
	Expect(mgr.GetFieldIndexer().IndexField(ctx, &corev1.Namespace{}, "metadata.name", nameIndexFunc)).To(Succeed())
	Expect(mgr.GetFieldIndexer().IndexField(ctx, &apiv1.Environment{}, "spec.preferredBranch", preferredBranchIndexFunc)).To(Succeed())
	Expect(k8s.AddOwnershipIndex(ctx, mgr.GetFieldIndexer(), &apiv1.Environment{})).To(Succeed())
	Expect(k8s.AddOwnershipIndex(ctx, mgr.GetFieldIndexer(), &apiv1.Deployment{})).To(Succeed())

	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(ctx)).To(Succeed())
	}()

	Eventually(
		func(g Gomega) { g.Expect(mgr.GetClient().List(ctx, &corev1.NamespaceList{})).To(Succeed()) },
		"30s",
	).Should(Succeed())

	return mgr.GetClient(), mgr.GetConfig()
}
