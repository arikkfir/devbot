package deployment_test

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"testing"

	_ "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var scheme *runtime.Scheme

var _ = BeforeSuite(func() {
	scheme = runtime.NewScheme()
	utilruntime.Must(apiv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
})

var _ = AfterSuite(func() {
	scheme = nil
})

func TestDeployment(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deployment")
}
