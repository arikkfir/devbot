package testing

import (
	"os"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/arikkfir/devbot/backend/internal/util/logging"
)

func init() {
	logging.Configure(os.Stderr, false, "trace", "0.0.0-local+unknown")
	logrLogger := logr.New(&logging.ZeroLogLogrAdapter{}).V(0)
	ctrl.SetLogger(logrLogger)
	klog.SetLogger(logrLogger)
}
