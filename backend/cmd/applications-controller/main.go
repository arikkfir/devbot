package main

import (
	"github.com/arikkfir/devbot/backend/internal/controllers"
	"github.com/arikkfir/devbot/backend/internal/controllers/application"
	"github.com/arikkfir/devbot/backend/internal/util"
	"github.com/arikkfir/devbot/backend/internal/util/initialization"
	"github.com/go-logr/logr"
	"github.com/rs/zerolog/log"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
)

var (
	cfg    controllers.ControllerConfig
	scheme = runtime.NewScheme()
)

func init() {
	initialization.Initialize(&cfg)
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))
}

func main() {
	ctrl.SetLogger(logr.New(&util.ZeroLogLogrAdapter{}))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                        scheme,
		Metrics:                       metricsserver.Options{BindAddress: cfg.MetricsAddr},
		HealthProbeBindAddress:        cfg.HealthProbeAddr,
		LeaderElection:                cfg.EnableLeaderElection,
		LeaderElectionID:              "f54ce4c0.devbot.kfirs.com",
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create manager")
	}

	mgrScheme := mgr.GetScheme()
	mgrClient := mgr.GetClient()

	applicationReconciler := &application.ApplicationReconciler{Client: mgrClient, Scheme: mgrScheme}
	if err := applicationReconciler.SetupWithManager(mgr); err != nil {
		log.Fatal().Err(err).Msg("Unable to create application controller")
	}

	applicationEnvReconciler := &application.ApplicationEnvironmentReconciler{Client: mgrClient, Scheme: mgrScheme}
	if err := applicationEnvReconciler.SetupWithManager(mgr); err != nil {
		log.Fatal().Err(err).Msg("Unable to create application environment controller")
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
