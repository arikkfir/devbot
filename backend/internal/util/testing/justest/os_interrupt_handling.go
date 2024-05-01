package justest

import (
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"os"
	"os/signal"
)

var (
	interruptionSignal os.Signal = nil
)

func init() {
	logging.Configure(os.Stderr, true, "trace")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	go func() {
		for s := range signalChan {
			interruptionSignal = s
			break
		}
	}()
}

//go:noinline
func VerifyNotInterrupted(t T) {
	GetHelper(t).Helper()
	if interruptionSignal != nil {
		t.Fatalf("Process has been canceled via signal %s", interruptionSignal)
	}
}
