package logging

import (
	"cmp"
	"context"
	"fmt"
	"github.com/arikkfir/command"
	"time"

	"github.com/getsentry/sentry-go"
)

type SentryInitHook struct {
	DSN string `desc:"Sentry DSN"`
}

func (h *SentryInitHook) PreRun(ctx context.Context) error {
	if h.DSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: h.DSN,
			// Set TracesSampleRate to 1.0 to capture 100%
			// of transactions for performance monitoring.
			// We recommend adjusting this value in production,
			TracesSampleRate: 1.0,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize Sentry: %w", err)
		}
	}
	return nil
}

type SentryFlushHook struct {
	Timeout time.Duration
}

func (h *SentryFlushHook) PostRun(context.Context, error, command.ExitCode) error {
	sentry.Flush(cmp.Or(h.Timeout, 2*time.Second))
	return nil
}
