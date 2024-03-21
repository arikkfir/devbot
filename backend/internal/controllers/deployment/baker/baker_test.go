package baker_test

import (
	"context"
	"embed"
	_ "embed"
	"github.com/arikkfir/devbot/backend/internal/controllers/deployment/baker"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"os"
	"path/filepath"
	"testing"
)

var (
	//go:embed expected-resources.yaml embed/*
	embeddedFS embed.FS
)

func TestBakerGenerateManifest(t *testing.T) {
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

	const (
		appName         = "myApp"
		preferredBranch = "my-feature"
		actualBranch    = "main"
		sha             = "abc123"
	)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	b := baker.NewBaker("kustomize", "yq")

	manifestFile, err := b.GenerateManifest(ctx, dir, appName, preferredBranch, actualBranch, sha)
	For(t).Expect(err).Will(BeNil())

	manifest, err := os.ReadFile(manifestFile)
	For(t).Expect(err).Will(BeNil())

	expectedBytes, err := embeddedFS.ReadFile("expected-resources.yaml")
	For(t).Expect(err).Will(BeNil())
	For(t).Expect(string(expectedBytes)).Will(BeEqualTo(string(manifest)))
}
