package e2e

import (
	"github.com/arikkfir/devbot/backend/internal/util/initialization"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	initialization.InitializeLogging(true, "info")

	smeeLogger := log.Level(zerolog.DebugLevel)
	cmd := exec.Command("smee", "--port", "8080", "--path", "/github/webhook")
	cmd.Stdout = smeeLogger
	cmd.Stderr = smeeLogger
	if err := cmd.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start smee")
	}

	exitCode := m.Run()

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		log.Fatal().Err(err).Msg("Failed to interrupt smee")
	} else if err := cmd.Wait(); err != nil {
		log.Fatal().Err(err).Msg("Failed to wait for smee to exit")
	}

	os.Exit(exitCode)
}
