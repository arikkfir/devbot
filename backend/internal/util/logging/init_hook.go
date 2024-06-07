package logging

import (
	"context"
	"github.com/arikkfir/devbot/backend/internal/util/version"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
)

// InitHook is the command executor for the "devbot" command.
type InitHook struct {
	DisableJSONLogging bool   `desc:"Disable JSON logging."`
	LogLevel           string `required:"true" desc:"Log level, must be one of: trace,debug,info,warn,error,fatal,panic"`
}

func (c *InitHook) PreRun(_ context.Context) error {
	Configure(os.Stderr, !c.DisableJSONLogging, c.LogLevel, version.Version)
	logrLogger := logr.New(&ZeroLogLogrAdapter{}).V(0)
	ctrl.SetLogger(logrLogger)
	klog.SetLogger(logrLogger)
	return nil
}
