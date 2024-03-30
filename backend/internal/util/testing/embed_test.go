package testing_test

import (
	"embed"
	_ "embed"
	internaltesting "github.com/arikkfir/devbot/backend/internal/util/testing"
	"github.com/secureworks/errors"
	"slices"
	"testing"
)

var (
	//go:embed all:embed/*
	embeddedFS embed.FS
)

func TestTraverseEmbeddedPath(t *testing.T) {
	testCases := []struct {
		name    string
		path    string
		handler func(string, []byte) error
		wantErr bool
	}{
		{
			name: "path is not part of handler path parameter",
			path: "embed/bare",
			handler: func(path string, data []byte) error {
				validPaths := []string{"embed/bare/README.md", "embed/bare/.test/test.yaml"}
				if slices.Contains(validPaths, path) {
					return nil
				} else {
					return errors.New("unexpected path '%s' - must be one of: %v", path, validPaths)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := internaltesting.TraverseEmbeddedPath(embeddedFS, tt.path, tt.handler); (err != nil) != tt.wantErr {
				t.Errorf("TraverseEmbeddedPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
