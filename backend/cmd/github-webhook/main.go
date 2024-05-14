package main

import (
	"context"

	"github.com/arikkfir/command"
	"github.com/go-logr/logr"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/arikkfir/devbot/backend/internal/webhooks/github"
	webhooksutil "github.com/arikkfir/devbot/backend/internal/webhooks/util"

	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// Version represents the version of the server. This variable gets its value by injection from the build process.
//
//goland:noinspection GoUnusedGlobalVariable
var Version = "0.0.0-unknown"

var rootCommand = command.New(command.Spec{
	Name:             filepath.Base(os.Args[0]),
	ShortDescription: "Devbot GitHub webhook connects GitHub events to Devbot installations.",
	LongDescription:  `This webhook will receive events from GitHub and mark the corresponding GitHub repository accordingly.'`,
	Config: &github.WebhookConfig{
		DisableJSONLogging: false,
		LogLevel:           "info",
		HealthPort:         9000,
		ServerPort:         8000,
	},
	Run: func(ctx context.Context, configAsAny any, usagePrinter command.UsagePrinter) error {
		cfg := configAsAny.(*github.WebhookConfig)

		// Configure logging
		logging.Configure(os.Stderr, !cfg.DisableJSONLogging, cfg.LogLevel, Version)
		logrLogger := logr.New(&logging.ZeroLogLogrAdapter{}).V(0)
		ctrl.SetLogger(logrLogger)
		klog.SetLogger(logrLogger)

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
			return errors.New("failed to create Kubernetes client: %w", err)
		}

		// Setup push handler
		handler, err := github.NewPushHandler(kubeConfig, scheme, cfg.WebhookSecret)
		if err != nil {
			return errors.New("failed to create push handler: %w", err)
		}
		if err := handler.Start(ctx); err != nil {
			return errors.New("failed to start push handler: %w", err)
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

		return nil
	},
})

func main() {
	command.Execute(rootCommand, os.Args, command.EnvVarsArrayToMap(os.Environ()))
}
