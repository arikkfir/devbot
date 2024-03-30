package main

import (
	"context"
	"github.com/arikkfir/devbot/backend/internal/controllers"
	"github.com/arikkfir/devbot/backend/internal/controllers/application"
	"github.com/arikkfir/devbot/backend/internal/util/configuration"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
)

var (
	cfg    controllers.Config
	scheme = runtime.NewScheme()
)

func init() {
	configuration.Parse(&cfg)
	logging.Configure(os.Stderr, cfg.DevMode, cfg.LogLevel)
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))
}

func main() {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                        scheme,
		Metrics:                       metricsserver.Options{BindAddress: cfg.MetricsAddr},
		HealthProbeBindAddress:        cfg.HealthProbeAddr,
		LeaderElection:                cfg.EnableLeaderElection,
		LeaderElectionID:              "f54ce4c2.devbot.kfirs.com",
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create manager")
	}

	mgrScheme := mgr.GetScheme()
	mgrClient := mgr.GetClient()

	if err := k8s.AddOwnershipIndex(context.Background(), mgr.GetFieldIndexer(), &apiv1.Environment{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to create index")
	}

	applicationReconciler := &application.Reconciler{Client: mgrClient, Config: cfg.CommandConfig, Scheme: mgrScheme}
	if err := applicationReconciler.SetupWithManager(mgr); err != nil {
		log.Fatal().Err(err).Msg("Unable to create application controller")
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatal().Err(err).Msg("Unable to set up health check")
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatal().Err(err).Msg("Unable to set up readiness check")
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatal().Err(err).Msg("Unable to run manager")
	}
}
