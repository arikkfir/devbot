package main

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/configuration"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/arikkfir/devbot/backend/internal/webhooks/github"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
)

var (
	cfg    github.WebhookConfig
	scheme = runtime.NewScheme()
)

func init() {
	configuration.Parse(&cfg)
	logging.Configure(os.Stderr, cfg.DevMode, cfg.LogLevel)
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))
}

func main() {
	// ctx := context.Background()
	//
	// // Create Kubernetes config
	// kubeConfig, err := rest.InClusterConfig()
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	// }
}
