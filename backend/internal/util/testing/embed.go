package testing

import (
	"embed"
	_ "embed"
	"github.com/secureworks/errors"
)

var (
	//go:embed all:embed/*
	embeddedFS embed.FS
)

func TraverseEmbeddedPath(path string, handler func(path string, data []byte) error) error {
	entries, err := embeddedFS.ReadDir("embed/" + path)
	if err != nil {
		return errors.New("failed to read embedded directory '%s': %w", path, err)
	}

	for _, entry := range entries {
		entryPath := path + "/" + entry.Name()
		if entry.IsDir() {
			if err := TraverseEmbeddedPath(entryPath, handler); err != nil {
				return err
			}
		} else if data, err := embeddedFS.ReadFile("embed/" + entryPath); err != nil {
			return errors.New("failed to read embedded file '%s': %w", entryPath, err)
		} else if err := handler(entryPath, data); err != nil {
			return errors.New("failed to process embedded file '%s': %w", entryPath, err)
		}
	}

	return nil
}
