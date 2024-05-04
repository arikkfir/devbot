package testing

import (
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
)

func init() {
	logging.Configure(os.Stderr, false, "trace", "0.0.0-local+unknown")
	logrLogger := logr.New(&logging.ZeroLogLogrAdapter{}).V(0)
	ctrl.SetLogger(logrLogger)
	klog.SetLogger(logrLogger)
}
