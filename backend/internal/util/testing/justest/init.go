package justest

import (
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"os"
	"os/signal"
	"testing"
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

func verifyNotInterruptedOnRootT(t JustT) {
	if interruptionSignal != nil {
		helper := GetHelper(t)
		rootT := helper.(*testing.T)
		if !rootT.Failed() {
			rootT.Fatalf("Process has been canceled via signal %s", interruptionSignal)
		}
	}
}

func verifyNotInterrupted(t JustT) {
	GetHelper(t).Helper()
	if interruptionSignal != nil {
		t.Fatalf("Process has been canceled via signal %s", interruptionSignal)
	}
}
