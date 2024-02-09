package e2e

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/format"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	scheme *runtime.Scheme
)

func init() {
	logging.Configure(ginkgo.GinkgoWriter, true, "trace")

	scheme = runtime.NewScheme()
	utilruntime.Must(apiv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	format.RegisterCustomFormatter(func(value interface{}) (string, bool) {
		if t, ok := value.(metav1.Time); ok {
			return t.String(), true
		}
		return "", false
	})
}
