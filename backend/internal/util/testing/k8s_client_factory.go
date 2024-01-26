package testing

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateKubernetesClient(scheme *runtime.Scheme, Client *client.WithWatch) {
	userHomeDir, err := os.UserHomeDir()
	Expect(err).ToNot(HaveOccurred())

	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	Expect(err).ToNot(HaveOccurred())

	c, err := client.NewWithWatch(kubeConfig, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	*Client = c

	DeferCleanup(func() { *Client = nil })
}
