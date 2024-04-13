package e2e

import (
	"context"
	v1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"testing"
	"time"
)

func TestTest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %+v", err)
	}

	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		t.Fatalf("Failed to build kubeconfig: %+v", err)
	}

	scheme := k8s.CreateScheme()
	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme: scheme,
		Client: client.Options{
			Scheme: scheme,
			Cache: &client.CacheOptions{
				DisableFor: nil,
			},
		},
		LeaderElection:         false,
		Metrics:                metricsserver.Options{BindAddress: "0"},
		HealthProbeBindAddress: "0",
		PprofBindAddress:       "0",
	})
	if err != nil {
		t.Fatalf("Failed to create manager: %+v", err)
	}

	go func() {
		if err := mgr.Start(For(t).Context()); err != nil {
			t.Logf("Kubernetes manager failed: %+v", err)
		}
	}()

	time.Sleep(10 * time.Second)

	apps := &v1.ApplicationList{}
	if err := mgr.GetClient().List(ctx, apps, &client.ListOptions{}); err != nil {
		t.Fatalf("Failed to list applications: %+v", err)
	}
}
