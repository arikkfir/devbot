package main

import (
	"github.com/arikkfir/devbot/backend/internal/controllers"
	"github.com/arikkfir/devbot/backend/internal/controllers/repositories/github"
	"github.com/arikkfir/devbot/backend/internal/util"
	"github.com/arikkfir/devbot/backend/internal/util/initialization"
	"github.com/go-logr/logr"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
		Scheme: scheme,
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor: []client.Object{
					// disable caching of secrets, as we might not get a "list" permission for them, and the default
					// cache tries to list objects for caching...
					&v1.Secret{},
				},
			},
		},
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

	githubRepositoryReconciler := &github.RepositoryReconciler{Client: mgrClient, Scheme: mgrScheme}
	if err := githubRepositoryReconciler.SetupWithManager(mgr); err != nil {
		log.Fatal().Err(err).Msg("Unable to create GitHub repository controller")
	}

	githubRepositoryRefReconciler := &github.RepositoryRefReconciler{Client: mgrClient, Scheme: mgrScheme}
	if err := githubRepositoryRefReconciler.SetupWithManager(mgr); err != nil {
		log.Fatal().Err(err).Msg("Unable to create GitHub repository ref controller")
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
