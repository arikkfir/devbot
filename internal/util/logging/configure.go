package logging

import (
	"io"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Configure(w io.Writer, jsonLogging bool, logLevel, version string) {

	// Add our version to the default logger
	log.Logger = log.With().Str("version", version).Logger()

	// Print stacktrace if one is available
	log.Logger = log.Logger.With().Stack().Logger()

	// If JSON logging is disabled, apply a human-friendly writer to the root logger
	if !jsonLogging {
		log.Logger = log.Logger.Output(zerolog.ConsoleWriter{Out: w})
	}

	// Set the default logger as the one returned for context.Context if one has not been associated with it yet
	zerolog.DefaultContextLogger = &log.Logger

	// Set the level of the root logger
	if level, err := zerolog.ParseLevel(logLevel); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config")
	} else {
		zerolog.SetGlobalLevel(level)
	}
}
