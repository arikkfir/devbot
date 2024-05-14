package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/arikkfir/command"
	"github.com/go-logr/logr"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	"github.com/arikkfir/devbot/backend/internal/controller"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/logging"

	"sigs.k8s.io/controller-runtime/pkg/client"

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

// Version represents the version of the controller. This variable gets its value by injection from the build process.
//
//goland:noinspection GoUnusedGlobalVariable
var Version = "0.0.0-unknown"

var rootCommand = command.New(command.Spec{
	Name:             filepath.Base(os.Args[0]),
	ShortDescription: "Devbot Controller runs the Kubernetes reconcilers.",
	LongDescription: `This controller runs the Kubernetes reconcilers that are in charge of continually reconciling
applications' desired state into an actual state in a Kubernetes cluster. It is responsible for managing the lifecycle
of repositories, applications, environments, and deployments.'`,
	Config: &controller.Config{
		DisableJSONLogging:   false,
		LogLevel:             "info",
		MetricsAddr:          ":8000",
		HealthProbeAddr:      ":9000",
		EnableLeaderElection: false,
		ContainerName:        "controller",
	},
	Run: func(ctx context.Context, configAsAny any, usagePrinter command.UsagePrinter) error {
		cfg := configAsAny.(*controller.Config)

		// Configure logging
		logging.Configure(os.Stderr, !cfg.DisableJSONLogging, cfg.LogLevel, Version)
		logrLogger := logr.New(&logging.ZeroLogLogrAdapter{}).V(0)
		ctrl.SetLogger(logrLogger)
		klog.SetLogger(logrLogger)

		// Create & register CRD scheme
		scheme := runtime.NewScheme()
		utilruntime.Must(clientgoscheme.AddToScheme(scheme))
		utilruntime.Must(apiv1.AddToScheme(scheme))

		// Create controller manager
		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme: scheme,
			Client: client.Options{
				Cache: &client.CacheOptions{
					DisableFor: []client.Object{
						// Disable caching of pods, since we only "get" our own pod, nothing more
						&v1.Pod{},

						// disable caching of secrets, as we might not get a "list" permission for them, and the default
						// cache tries to list objects for caching...
						&v1.Secret{},
					},
				},
			},
			Metrics:                       metricsserver.Options{BindAddress: cfg.MetricsAddr},
			HealthProbeBindAddress:        cfg.HealthProbeAddr,
			LeaderElection:                cfg.EnableLeaderElection,
			LeaderElectionID:              "f54ce4c2.devbot.kfirs.com",
			LeaderElectionReleaseOnCancel: true,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to create manager")
		}

		// Create indices used by the controllers
		if err := k8s.AddOwnershipIndex(ctx, mgr.GetFieldIndexer(), &batchv1.Job{}); err != nil {
			log.Fatal().Err(err).Msg("Failed to create job index")
		}
		if err := k8s.AddOwnershipIndex(ctx, mgr.GetFieldIndexer(), &apiv1.Environment{}); err != nil {
			log.Fatal().Err(err).Msg("Failed to create environment index")
		}
		if err := k8s.AddOwnershipIndex(ctx, mgr.GetFieldIndexer(), &apiv1.Deployment{}); err != nil {
			log.Fatal().Err(err).Msg("Failed to create deployment index")
		}

		// Create & register application controller
		repositoryReconciler := &controller.RepositoryReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
		if err := repositoryReconciler.SetupWithManager(mgr); err != nil {
			log.Fatal().Err(err).Msg("Unable to create repository controller")
		}

		// Create & register application controller
		applicationReconciler := &controller.ApplicationReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
		if err := applicationReconciler.SetupWithManager(mgr); err != nil {
			log.Fatal().Err(err).Msg("Unable to create application controller")
		}

		// Create & register environment controller
		environmentReconciler := &controller.EnvironmentReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
		if err := environmentReconciler.SetupWithManager(mgr); err != nil {
			log.Fatal().Err(err).Msg("Unable to create environment controller")
		}

		// Create & register environment controller
		deploymentReconciler := &controller.DeploymentReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Config: *cfg}
		if err := deploymentReconciler.SetupWithManager(mgr); err != nil {
			log.Fatal().Err(err).Msg("Unable to create deployment controller")
		}

		// Add health & readiness checks
		if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
			log.Fatal().Err(err).Msg("Unable to set up health check")
		}
		if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
			log.Fatal().Err(err).Msg("Unable to set up readiness check")
		}

		// Start the controller manager
		if err := mgr.Start(ctx); err != nil {
			log.Fatal().Err(err).Msg("Unable to run manager")
		}

		return nil
	},
})

func main() {
	command.Execute(rootCommand, os.Args, command.EnvVarsArrayToMap(os.Environ()))
}
