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
	"github.com/arikkfir/devbot/backend/internal/util/version"
	"github.com/arikkfir/devbot/backend/internal/webhooks/github"
	webhooksutil "github.com/arikkfir/devbot/backend/internal/webhooks/util"

	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type Executor struct {
	DisableJSONLogging bool   `desc:"Disable JSON logging."`
	LogLevel           string `required:"true" desc:"Log level, must be one of: trace,debug,info,warn,error,fatal,panic"`
	HealthPort         int    `required:"true" desc:"Health endpoint port."`
	ServerPort         int    `required:"true" desc:"Webhook server port."`
	WebhookSecret      string `required:"true" desc:"Webhook secret shared by this server and GitHub."`
}

func (e *Executor) PreRun(_ context.Context) error { return nil }
func (e *Executor) Run(ctx context.Context) error {

	// Configure logging
	logging.Configure(os.Stderr, !e.DisableJSONLogging, e.LogLevel, version.Version)
	logrLogger := logr.New(&logging.ZeroLogLogrAdapter{}).V(0)
	ctrl.SetLogger(logrLogger)
	klog.SetLogger(logrLogger)

	// Setup Kubernetes scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))

	// Setup health check
	hc := webhooksutil.NewHealthCheckServer(e.HealthPort)
	go hc.Start(ctx)
	defer hc.Stop(ctx)

	// Create Kubernetes config
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return errors.New("failed to create Kubernetes client: %w", err)
	}

	// Setup push handler
	handler, err := github.NewPushHandler(kubeConfig, scheme, e.WebhookSecret)
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
		Addr:    ":" + strconv.Itoa(e.ServerPort),
		Handler: webhooksutil.AccessLogMiddleware(false, nil, mux),
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Err(err).Msg("HTTP server failed")
	}

	return nil
}

func main() {

	// Create command structure
	cmd := command.MustNew(
		filepath.Base(os.Args[0]),
		"Devbot GitHub webhook connects GitHub events to Devbot installations.",
		`This webhook will receive events from GitHub and mark the corresponding GitHub repository accordingly.'`,
		&Executor{
			DisableJSONLogging: false,
			LogLevel:           "info",
			HealthPort:         8000,
			ServerPort:         9000,
		},
	)

	// Prepare a context that gets canceled if OS termination signals are sent
	ctx, cancel := context.WithCancel(command.SetupSignalHandler())
	defer cancel()

	// Execute the correct command
	command.Execute(ctx, os.Stderr, cmd, os.Args, command.EnvVarsArrayToMap(os.Environ()))

}
