package e2e

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func CreateKubernetesClient(scheme *runtime.Scheme, Client *client.Client) {
	userHomeDir, err := os.UserHomeDir()
	Expect(err).ToNot(HaveOccurred())

	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	Expect(err).ToNot(HaveOccurred())

	mgrContext, cancel := context.WithCancel(context.Background())
	DeferCleanup(cancel)

	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme: scheme,
		Cache: cache.Options{
			Scheme: scheme,
		},
		LeaderElection:         false,
		Metrics:                metricsserver.Options{BindAddress: "0"},
		HealthProbeBindAddress: "0",
		PprofBindAddress:       "0",
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8s.AddOwnershipIndex(mgrContext, mgr.GetCache(), &apiv1.GitHubRepositoryRef{})).To(Succeed())
	Expect(k8s.AddOwnershipIndex(mgrContext, mgr.GetCache(), &apiv1.Deployment{})).To(Succeed())
	Expect(k8s.AddOwnershipIndex(mgrContext, mgr.GetCache(), &apiv1.Environment{})).To(Succeed())

	go func(ctx context.Context) {
		if err := mgr.Start(ctx); err != nil {
			log.Fatal().Err(err).Msg("Kubernetes manager failed")
		} else {
			log.Info().Msg("Kubernetes manager stopped")
		}
	}(mgrContext)

	*Client = mgr.GetClient()
	DeferCleanup(func() { *Client = nil })
}
