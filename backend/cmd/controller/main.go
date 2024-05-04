package main

import (
	"github.com/arikkfir/devbot/backend/internal/controller"
	. "github.com/arikkfir/devbot/backend/internal/util/configuration"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/go-logr/logr"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"os"
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

const (
	disableJSONLoggingKey   = "disable-json-logging"
	logLevelKey             = "log-level"
	metricsAddrKey          = "metrics-bind-address"
	healthProbeAddrKey      = "health-probe-bind-address"
	enableLeaderElectionKey = "leader-elect"
	podNamespaceKey         = "pod-namespace"
	podNameKey              = "pod-name"
	containerNameKey        = "container-name"
)

// Version represents the version of the controller. This variable gets its value by injection from the build process.
//
//goland:noinspection GoUnusedGlobalVariable
var Version = "0.0.0-unknown"

// cfg is the configuration of the controller. It is populated in the init function.
var cfg = controller.Config{
	DisableJSONLogging:   false,
	LogLevel:             "info",
	MetricsAddr:          ":8000",
	HealthProbeAddr:      ":9000",
	EnableLeaderElection: false,
	ContainerName:        "controller",
}

func init() {

	// Configure & parse CLI flags
	pflag.BoolVar(&cfg.DisableJSONLogging, disableJSONLoggingKey, cfg.DisableJSONLogging, "Disable JSON logging")
	pflag.StringVar(&cfg.LogLevel, logLevelKey, cfg.LogLevel, "Log level, must be one of: trace,debug,info,warn,error,fatal,panic")
	pflag.StringVar(&cfg.MetricsAddr, metricsAddrKey, cfg.MetricsAddr, "Address the metric endpoint should bind to")
	pflag.StringVar(&cfg.HealthProbeAddr, healthProbeAddrKey, cfg.HealthProbeAddr, "Address the health endpoint should bind to")
	pflag.BoolVar(&cfg.EnableLeaderElection, enableLeaderElectionKey, cfg.EnableLeaderElection, "Enable leader election, ensuring only one controller is active")
	pflag.StringVar(&cfg.PodNamespace, podNamespaceKey, cfg.PodNamespace, "Namespace of the controller pod (usually provided via downward API)")
	pflag.StringVar(&cfg.PodName, podNameKey, cfg.PodName, "Name of the controller pod (usually provided via downward API)")
	pflag.StringVar(&cfg.ContainerName, containerNameKey, cfg.ContainerName, "Name of the controller container")
	pflag.Parse()

	// Allow the user to override configuration values using environment variables
	ApplyBoolEnvironmentVariableTo(&cfg.DisableJSONLogging, FlagNameToEnvironmentVariable(disableJSONLoggingKey))
	ApplyStringEnvironmentVariableTo(&cfg.LogLevel, FlagNameToEnvironmentVariable(logLevelKey))
	ApplyStringEnvironmentVariableTo(&cfg.MetricsAddr, FlagNameToEnvironmentVariable(metricsAddrKey))
	ApplyStringEnvironmentVariableTo(&cfg.HealthProbeAddr, FlagNameToEnvironmentVariable(healthProbeAddrKey))
	ApplyBoolEnvironmentVariableTo(&cfg.EnableLeaderElection, FlagNameToEnvironmentVariable(enableLeaderElectionKey))
	ApplyStringEnvironmentVariableTo(&cfg.PodNamespace, FlagNameToEnvironmentVariable(podNamespaceKey))
	ApplyStringEnvironmentVariableTo(&cfg.PodName, FlagNameToEnvironmentVariable(podNameKey))
	ApplyStringEnvironmentVariableTo(&cfg.ContainerName, FlagNameToEnvironmentVariable(containerNameKey))

	// Validate configuration
	if cfg.LogLevel == "" {
		log.Fatal().Msg("Log level cannot be empty")
	}
	if cfg.MetricsAddr == "" {
		log.Fatal().Msg("Metrics bind address cannot be empty")
	}
	if cfg.HealthProbeAddr == "" {
		log.Fatal().Msg("Health probe bind address cannot be empty")
	}
	if cfg.PodNamespace == "" {
		log.Fatal().Msg("Pod namespace cannot be empty")
	}
	if cfg.PodName == "" {
		log.Fatal().Msg("Pod name cannot be empty")
	}
	if cfg.ContainerName == "" {
		log.Fatal().Msg("Container name cannot be empty")
	}

	// Configure logging
	logging.Configure(os.Stderr, !cfg.DisableJSONLogging, cfg.LogLevel, Version)
	logrLogger := logr.New(&logging.ZeroLogLogrAdapter{}).V(0)
	ctrl.SetLogger(logrLogger)
	klog.SetLogger(logrLogger)
}

func main() {

	// Create a context that cancels when a signal is received
	ctx := ctrl.SetupSignalHandler()

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
	deploymentReconciler := &controller.DeploymentReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Config: cfg}
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
}
