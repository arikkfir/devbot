package e2e

import (
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"github.com/onsi/ginkgo/v2"
)

func init() {
	logging.Configure(ginkgo.GinkgoWriter, true, "trace")
}
