package applier_test

import (
	"context"
	"embed"
	_ "embed"
	"github.com/arikkfir/devbot/backend/internal/controllers/deployment/applier"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"os"
	"path/filepath"
	"testing"
)

var (
	//go:embed expected-resources.yaml embed/*
	embeddedFS embed.FS
)

func TestApplierApply(t *testing.T) {
	t.Skipf("Skipping until we can set up a mock cluster for 'kubectl' to connect to")

	// Create temp directory for extracting kustomization dir to
	dir, err := os.MkdirTemp("", t.Name())
	For(t).Expect(err).Will(BeNil())
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	// Extract embedded kustomization dir
	entries, err := embeddedFS.ReadDir("embed")
	For(t).Expect(err).Will(BeNil())
	for _, entry := range entries {
		bytes, err := embeddedFS.ReadFile("embed/" + entry.Name())
		For(t).Expect(err).Will(BeNil())

		targetPath := filepath.Join(dir, entry.Name())
		For(t).Expect(os.WriteFile(targetPath, bytes, 0644)).Will(Succeed())
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	a := applier.NewApplier("kubectl")
	For(t).Expect(a.Apply(ctx, filepath.Join(dir, "resources.yaml"))).Will(Succeed())
	// TODO: verify resources applied correctly
}
