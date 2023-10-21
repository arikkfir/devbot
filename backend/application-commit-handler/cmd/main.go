package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/arikkfir/devbot/backend/application-commit-handler/internal"
	appsv1 "github.com/arikkfir/devbot/backend/applications-controller/api/v1"
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"net/http"
	"os"
	"strconv"
)

var cfg internal.Config

func init() {
	parser := flags.NewParser(&cfg, flags.HelpFlag|flags.PassDoubleDash)

	if _, err := parser.Parse(); err != nil {
		fmt.Printf("ERROR: %s\n\n", err)
		parser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	if cfg.DevMode {
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
				return nil
			},
		}
		log.Logger = log.Output(writer)
		zerolog.DefaultContextLogger = &log.Logger
	} else {
		zerolog.ErrorStackMarshaler = func(err error) interface{} {
			buffer := bytes.Buffer{}
			errors.PrintStackChain(&buffer, err)
			return buffer.String()
		}
	}

	if level, err := zerolog.ParseLevel(cfg.LogLevel); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config")
	} else {
		zerolog.SetGlobalLevel(level)
	}

	zerolog.DefaultContextLogger = &log.Logger
}

func main() {
	ctx := context.Background()

	// Setup health check
	hc := internal.NewHealthCheckServer(cfg.HTTP.HealthPort)
	go hc.Start(ctx)
	defer hc.Stop(ctx)

	// Register used CRDs
	err := appsv1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register CRDs")
	}

	// Create Kubernetes client
	k8sClient, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Setup push handler
	handler := internal.NewPushHandler(k8sClient, cfg.WebhookSecret)

	// Setup server
	server := &http.Server{
		Addr: ":" + strconv.Itoa(cfg.HTTP.Port),
		Handler: internal.AccessLogMiddleware(
			cfg.HTTP.AccessLogExcludeRemoteAddr,
			cfg.HTTP.AccessLogExcludedHeaders,
			handler.Handler,
		),
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Err(err).Msg("HTTP server failed")
	}
}
