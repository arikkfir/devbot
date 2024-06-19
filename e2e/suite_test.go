package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/arikkfir/devbot/e2e/util"
)

var gh *github.Client
var c client.Client
var rc *rest.Config

func TestEndToEnd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "End-to-end Suite")
}

var _ = BeforeSuite(func(ctx context.Context) {
	ctrl.SetLogger(GinkgoLogr)
	klog.SetLogger(GinkgoLogr)

	format.MaxLength = 1024 * 32

	wd, err := os.Getwd()
	Expect(err).To(BeNil())
	for {
		e2ePath := filepath.Join(wd, "e2e.env")
		_, _ = fmt.Fprintf(GinkgoWriter, "Searching environment variables at: %s\n", e2ePath)
		if stat, err := os.Stat(e2ePath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				Fail(fmt.Sprintf("failed getting file information for '%s': %v", e2ePath, err))
			}
		} else if stat.IsDir() {
			Fail(fmt.Sprintf("'%s' is a directory", e2ePath))
		} else {
			_, _ = fmt.Fprintf(GinkgoWriter, "Reading environment variables from: %s\n", e2ePath)
			env, err := os.ReadFile(e2ePath)
			Expect(err).To(BeNil(), "failed reading '%s'", e2ePath)
			for _, line := range strings.Split(string(env), "\n") {
				if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
					// Skip empty lines and comments
					continue
				}

				tokens := strings.SplitN(line, "=", 2)
				Expect(tokens).To(HaveLen(2))
				Expect(os.Setenv(tokens[0], tokens[1])).To(Succeed(), "failed setting environment variable '%s' to '%s'", tokens[0], tokens[1])
			}
		}
		if filepath.Base(wd) == "devbot" {
			break
		} else {
			wd = filepath.Dir(wd)
		}
	}

	helmReleasePrefix := "devbot"
	if v, found := os.LookupEnv("DEVBOT_PREFIX"); found {
		helmReleasePrefix = v
	}
	util.DevbotControllerServiceAccountName = fmt.Sprintf("%s-controller", helmReleasePrefix)

	gh = util.NewGitHubClient(ctx)
	c, rc = util.NewK8sClient()
})
