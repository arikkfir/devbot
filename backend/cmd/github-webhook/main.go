package main

import (
	"context"
	"github.com/arikkfir/command"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/arikkfir/devbot/backend/internal/webhooks/github"
	webhooksutil "github.com/arikkfir/devbot/backend/internal/webhooks/util"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type Action struct {
	HealthPort int `required:"true" desc:"Health endpoint port."`
	ServerPort int `required:"true" desc:"Webhook server port."`
}

func (e *Action) Run(ctx context.Context) error {

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
	handler, err := github.NewPushHandler(kubeConfig, scheme)
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
		&Action{
			HealthPort: 8000,
			ServerPort: 9000,
		},
		[]command.PreRunHook{&logging.InitHook{LogLevel: "info"}, &logging.SentryInitHook{}},
		[]command.PostRunHook{&logging.SentryFlushHook{}},
	)

	// Prepare a context that gets canceled if OS termination signals are sent
	ctx, cancel := context.WithCancel(command.SetupSignalHandler())
	defer cancel()

	// Execute the correct command
	os.Exit(int(command.Execute(ctx, os.Stderr, cmd, os.Args, command.EnvVarsArrayToMap(os.Environ()))))

}
