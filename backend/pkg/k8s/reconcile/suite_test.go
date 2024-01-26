package reconcile_test

import (
	_ "github.com/arikkfir/devbot/backend/internal/util/testing"
	testingv1 "github.com/arikkfir/devbot/backend/internal/util/testing/api/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"testing"
)

var scheme *runtime.Scheme

var _ = BeforeSuite(func() {
	scheme = runtime.NewScheme()
	utilruntime.Must(testingv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
})

var _ = AfterSuite(func() {
	scheme = nil
})

func TestReconcile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg.k8s.reconcile")
}
