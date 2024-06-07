package devctl

import (
	"context"
	"os"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/arikkfir/devbot/backend/internal/util/logging"
)

// Version represents the version of the server. This variable gets its value by injection from the build process.
//
//goland:noinspection GoUnusedGlobalVariable
var Version = "0.0.0-unknown"

// RootExecutor is the command executor for the "devbot" command.
type RootExecutor struct {
	DisableJSONLogging bool
	LogLevel           string
}

func (c *RootExecutor) PreRun(ctx context.Context) error {
	logging.Configure(os.Stderr, c.DisableJSONLogging, c.LogLevel, Version)
	logrLogger := logr.New(&logging.ZeroLogLogrAdapter{}).V(0)
	ctrl.SetLogger(logrLogger)
	klog.SetLogger(logrLogger)
	return nil
}

func (c *RootExecutor) Run(ctx context.Context) error { return nil }
