package main

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/internal/util/configuration"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/arikkfir/devbot/backend/internal/webhooks/github"
	webhooksutil "github.com/arikkfir/devbot/backend/internal/webhooks/util"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"net/http"
	"os"
	"strconv"
)

const (
	disableJSONLoggingKey = "disable-json-logging"
	logLevelKey           = "log-level"
	healthPortKey         = "health-port"
	serverPortKey         = "server-port"
	webhookSecretKey      = "webhook-secret"
)

// Version represents the version of the server. This variable gets its value by injection from the build process.
//
//goland:noinspection GoUnusedGlobalVariable
var Version = "0.0.0-unknown"

// cfg is the configuration of the server. It is populated in the init function.
var cfg = github.WebhookConfig{
	DisableJSONLogging: false,
	LogLevel:           "info",
	HealthPort:         9000,
	ServerPort:         8000,
}

func init() {

	// Configure & parse CLI flags
	pflag.BoolVar(&cfg.DisableJSONLogging, disableJSONLoggingKey, cfg.DisableJSONLogging, "Disable JSON logging")
	pflag.StringVar(&cfg.LogLevel, logLevelKey, cfg.LogLevel, "Log level, must be one of: trace,debug,info,warn,error,fatal,panic")
	pflag.IntVar(&cfg.HealthPort, healthPortKey, cfg.HealthPort, "Port to listen on for health checks")
	pflag.IntVar(&cfg.ServerPort, serverPortKey, cfg.ServerPort, "Port to listen on")
	pflag.StringVar(&cfg.WebhookSecret, webhookSecretKey, cfg.WebhookSecret, "Webhook secret")
	pflag.Parse()

	// Allow the user to override configuration values using environment variables
	ApplyBoolEnvironmentVariableTo(&cfg.DisableJSONLogging, FlagNameToEnvironmentVariable(disableJSONLoggingKey))
	ApplyStringEnvironmentVariableTo(&cfg.LogLevel, FlagNameToEnvironmentVariable(logLevelKey))
	ApplyIntEnvironmentVariableTo(&cfg.HealthPort, FlagNameToEnvironmentVariable(healthPortKey))
	ApplyIntEnvironmentVariableTo(&cfg.ServerPort, FlagNameToEnvironmentVariable(serverPortKey))
	ApplyStringEnvironmentVariableTo(&cfg.WebhookSecret, FlagNameToEnvironmentVariable(webhookSecretKey))

	// Validate configuration
	if cfg.LogLevel == "" {
		log.Fatal().Msg("Log level cannot be empty")
	}
	if cfg.HealthPort == 0 {
		log.Fatal().Msg("Health port cannot be zero")
	}
	if cfg.ServerPort == 0 {
		log.Fatal().Msg("Server port cannot be zero")
	}
	if cfg.WebhookSecret == "" {
		log.Fatal().Msg("Webhook secret cannot be empty")
	}

	// Configure logging
	logging.Configure(os.Stderr, !cfg.DisableJSONLogging, cfg.LogLevel, Version)
}

func main() {
	ctx := context.Background()

	// Setup Kubernetes scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))

	// Setup health check
	hc := webhooksutil.NewHealthCheckServer(cfg.HealthPort)
	go hc.Start(ctx)
	defer hc.Stop(ctx)

	// Create Kubernetes config
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Setup push handler
	handler, err := github.NewPushHandler(kubeConfig, scheme, cfg.WebhookSecret)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create push handler")
	}
	if err := handler.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start push handler")
	}
	defer func(handler *github.PushHandler) {
		err := handler.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close push handler")
		}
	}(handler)

	// Setup routing
	mux := http.NewServeMux()
	mux.HandleFunc("/github/webhook", handler.HandleWebhookRequest)

	// Setup server
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.ServerPort),
		Handler: webhooksutil.AccessLogMiddleware(false, nil, mux),
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Err(err).Msg("HTTP server failed")
	}
}
