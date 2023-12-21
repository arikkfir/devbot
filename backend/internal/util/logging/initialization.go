package logging

import (
	"bytes"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"io"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
)

func ErrorStackMarshaler(err error) interface{} {
	// TODO: add metadata & tags from event error, if any
	buffer := bytes.Buffer{}
	errors.PrintStackChain(&buffer, err)
	return buffer.String()
}

func Configure(w io.Writer, devMode bool, logLevel string) {
	if devMode {
		// Set an error stack marshaller which simply prints the stack trace as a string
		// This string will be used afterward by the "FormatExtra" to print it nicely AFTER
		// the log message line
		// This creates a similar effect to Java & Python log output experience
		zerolog.ErrorStackMarshaler = ErrorStackMarshaler
		writer := zerolog.ConsoleWriter{
			Out:           w,
			FieldsExclude: []string{zerolog.ErrorStackFieldName},
			FormatExtra: func(event map[string]interface{}, b *bytes.Buffer) error {
				stack, ok := event[zerolog.ErrorStackFieldName]
				if ok {
					stackString := stack.(string)
					//indentedStackString := text.Indent(stackString, "     ")
					_, err := fmt.Fprintf(b, "\n%5s", stackString)
					if err != nil {
						panic(err)
					}
				}
				return nil
			},
		}
		log.Logger = log.Output(writer).With().Stack().Logger()
		zerolog.DefaultContextLogger = &log.Logger
	} else {
		zerolog.ErrorStackMarshaler = ErrorStackMarshaler
		log.Logger = log.With().Stack().Logger()
	}

	if level, err := zerolog.ParseLevel(logLevel); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config")
	} else {
		log.Info().Msgf("Configured log level to %s", strings.ToUpper(level.String()))
		zerolog.SetGlobalLevel(level)
	}

	zerolog.DefaultContextLogger = &log.Logger

	logrLogger := logr.New(&zeroLogLogrAdapter{}).V(0)
	ctrl.SetLogger(logrLogger)
}
