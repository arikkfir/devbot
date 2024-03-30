package logging

import (
	"bytes"
	"context"
	"fmt"
	"github.com/arikkfir/devbot/backend/internal/config"
	"github.com/distribution/reference"
	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"io"
	corev1 "k8s.io/api/core/v1"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		zerolog.SetGlobalLevel(level)
	}

	zerolog.DefaultContextLogger = &log.Logger

	config.PodName = os.Getenv("POD_NAME")
	config.PodNamespace = os.Getenv("POD_NAMESPACE")
	if config.PodName != "" && config.PodNamespace != "" {
		c, err := client.New(ctrl.GetConfigOrDie(), client.Options{})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create in-cluster Kubernetes client")
		}
		pod := &corev1.Pod{}
		podObjectKey := client.ObjectKey{Namespace: config.PodNamespace, Name: config.PodName}
		if err := c.Get(context.Background(), podObjectKey, pod); err != nil {
			log.Fatal().Err(err).Msg("Failed to get pod")
		}
		config.Image = pod.Spec.Containers[0].Image
		r, err := reference.Parse(pod.Spec.Containers[0].Image)
		if tagged, ok := r.(reference.Tagged); ok {
			config.Version = tagged.Tag()
		} else {
			log.Info().Msg("Version could not be auto-configured (pod image is not tagged)")
		}
	} else {
		log.Info().Msg("Version could not be auto-configured (POD_NAME or POD_NAMESPACE not set)")
	}
	log.Logger = log.With().Str("version", config.Version).Logger()

	logrLogger := logr.New(&zeroLogLogrAdapter{}).V(0)
	ctrl.SetLogger(logrLogger)
}
