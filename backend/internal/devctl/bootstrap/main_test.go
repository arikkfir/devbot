package bootstrap

import (
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"

	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
)

var testEnv env.Environment

func TestMain(m *testing.M) {
	kindClusterName := stringsutil.Name()

	testEnv = env.New()
	testEnv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
	)
	testEnv.Finish(envfuncs.DestroyCluster(kindClusterName))

	os.Exit(testEnv.Run(m))
}
