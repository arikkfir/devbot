package testing

import (
	"embed"
	"github.com/secureworks/errors"
)

func TraverseEmbeddedPath(fs embed.FS, path string, handler func(path string, data []byte) error) error {
	entries, err := fs.ReadDir(path)
	if err != nil {
		return errors.New("failed to read embedded directory '%s': %w", path, err)
	}

	for _, entry := range entries {
		entryPath := path + "/" + entry.Name()
		if entry.IsDir() {
			if err := TraverseEmbeddedPath(fs, entryPath, handler); err != nil {
				return err
			}
		} else if data, err := fs.ReadFile(entryPath); err != nil {
			return errors.New("failed to read embedded file '%s': %w", entryPath, err)
		} else if err := handler(entryPath, data); err != nil {
			return errors.New("failed to process embedded file '%s': %w", entryPath, err)
		}
	}

	return nil
}
