package e2e

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/initialization"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	initialization.InitializeLogging(true, "debug")

	// Register our CRDs
	err := apiv1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register CRDs")
	}

	os.Exit(m.Run())
}
