package initialization

import (
	"bytes"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"os"
)

type Config interface {
	IsDevMode() bool
	GetLogLevel() string
}

func InitializeLogging(devMode bool, logLevel string) {
	if devMode {
		// Set an error stack marshaller which simply prints the stack trace as a string
		// This string will be used afterward by the "FormatExtra" to print it nicely AFTER
		// the log message line
		// This creates a similar effect to Java & Python log output experience
		zerolog.ErrorStackMarshaler = func(err error) interface{} {
			buffer := bytes.Buffer{}
			errors.PrintStackChain(&buffer, err)
			return buffer.String()
		}
		writer := zerolog.ConsoleWriter{
			Out:           os.Stderr,
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
				// TODO: add metadata & tags from event error, if any
				return nil
			},
		}
		log.Logger = log.Output(writer).With().Stack().Logger()
		zerolog.DefaultContextLogger = &log.Logger
	} else {
		zerolog.ErrorStackMarshaler = func(err error) interface{} {
			buffer := bytes.Buffer{}
			errors.PrintStackChain(&buffer, err)
			return buffer.String()
		}
		// TODO: add metadata & tags from event error, if any
		log.Logger = log.With().Stack().Logger()
	}

	if level, err := zerolog.ParseLevel(logLevel); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config")
	} else {
		zerolog.SetGlobalLevel(level)
	}

	zerolog.DefaultContextLogger = &log.Logger
}

func Initialize(cfg Config) {
	parser := flags.NewParser(cfg, flags.HelpFlag|flags.PassDoubleDash)
	if _, err := parser.Parse(); err != nil {
		fmt.Printf("ERROR: %s\n\n", err)
		parser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	InitializeLogging(cfg.IsDevMode(), cfg.GetLogLevel())
}
