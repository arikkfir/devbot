package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/alexliesenfeld/health"
	"github.com/arikkfir/devbot/api/internal/config"
	"github.com/arikkfir/devbot/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"time"
)

type API struct {
	Log  config.LogConfig  `group:"log" namespace:"log" env-namespace:"LOG" description:"Logging configuration"`
	HTTP config.HTTPConfig `group:"http" namespace:"http" env-namespace:"HTTP" description:"HTTP configuration"`
}

func main() {
	cfg := API{}

	// Read configuration
	parser := flags.NewParser(&cfg, flags.Default)
	parser.NamespaceDelimiter = "-"
	_, err := parser.Parse()
	if err != nil {
		if flags.WroteHelp(err) {
			os.Exit(0)
			return
		} else {
			logrus.WithError(err).Fatal("Configuration error")
		}
	}

	// Initialize logging
	logger := logrus.StandardLogger()
	logger.SetOutput(os.Stdout)
	logger.SetReportCaller(cfg.Log.CallerInfo)
	if cfg.Log.JSONLogging {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat:   time.RFC3339,
			DisableTimestamp:  false,
			DisableHTMLEscape: true,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "severity",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
			PrettyPrint: false,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        time.RFC3339,
			DisableLevelTruncation: true,
			PadLevelText:           true,
		})
	}

	// Redirect os.Stdout and os.Stderr to logging framework
	infoWriter := logger.WriterLevel(logrus.InfoLevel)
	defer infoWriter.Close()
	warnWriter := logger.WriterLevel(logrus.ErrorLevel)
	defer warnWriter.Close()
	log.SetOutput(infoWriter)

	// Redirect Gin logging to logging framework
	gin.DefaultWriter = infoWriter
	gin.DefaultErrorWriter = warnWriter

	// Configured!
	logrus.WithField("config", fmt.Sprintf("%+v", &cfg)).Info("Configured")

	// Setup web server
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RequestLoggerFactory(cfg.HTTP.LogResponseBody))
	// TODO: r.Use(middleware.Recovery)
	// TODO: r.Use(middleware.ErrorHandler)
	r.Get("/health", health.NewHandler(health.NewChecker(
		health.WithCacheDuration(1*time.Second),
		health.WithTimeout(10*time.Second),
		health.WithCheck(health.Check{
			Name:  "up",
			Check: func(ctx context.Context) error { return nil },
		}),
		health.WithPeriodicCheck(15*time.Second, 3*time.Second, health.Check{
			Name:  "heavy-up",
			Check: func(ctx context.Context) error { return nil },
		}),
		health.WithStatusListener(func(ctx context.Context, state health.CheckerState) {
			logrus.
				WithField("checkers", state.CheckState).
				WithField("status", state.Status).
				Warn("Health status changed")
		}),
	)))

	// Set up the routes
	r.MethodFunc("GET", "/github/XXX", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	// Start the server
	err = http.ListenAndServe(cfg.HTTP.Address, r)
	if !errors.Is(err, http.ErrServerClosed) {
		logrus.WithError(err).Fatal("HTTP server failed")
	}
}
