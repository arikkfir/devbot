package observability

import (
	"context"
	"os"

	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/rs/zerolog/log"

	"github.com/arikkfir/devbot/internal/util/version"
)

// LoggingHook is the command executor for the "devbot" command.
type LoggingHook struct {
	DisableJSONLogging bool   `desc:"Disable JSON logging."`
	LogLevel           string `required:"true" desc:"Log level, must be one of: trace,debug,info,warn,error,fatal,panic"`
}

func (c *LoggingHook) PreRun(_ context.Context) error {

	// Add our version to the default logger
	log.Logger = log.With().Str("version", version.Version).Logger()

	// Print stacktrace if one is available
	log.Logger = log.Logger.With().Stack().Logger()

	// If JSON logging is disabled, apply a human-friendly writer to the root logger
	if c.DisableJSONLogging {
		log.Logger = log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Set the default logger as the one returned for context.Context if one has not been associated with it yet
	zerolog.DefaultContextLogger = &log.Logger

	// Set the level of the root logger
	if level, err := zerolog.ParseLevel(c.LogLevel); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config")
	} else {
		zerolog.SetGlobalLevel(level)
	}

	ctrlRuntimeLogrToZerologAdapter := logr.New(&ZeroLogLogrAdapter{l: log.Logger})
	ctrl.SetLogger(ctrlRuntimeLogrToZerologAdapter)
	klog.SetLogger(ctrlRuntimeLogrToZerologAdapter)

	otelLogrToZerologAdapter := logr.New(&ZeroLogLogrAdapter{l: log.Logger.Level(zerolog.InfoLevel)})
	otel.SetLogger(otelLogrToZerologAdapter)

	log.Info().Msg("Logging configured")

	return nil
}
