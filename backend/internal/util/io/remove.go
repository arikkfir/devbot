package io

import (
	"github.com/secureworks/errors"
	"os"
	"path/filepath"
)

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return errors.New("failed opening directory: %w", err)
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return errors.New("failed reading directory names: %w", err)
	}

	for _, name := range names {
		fullName := filepath.Join(dir, name)
		err = os.RemoveAll(fullName)
		if err != nil {
			return errors.New("failed removing '%s': %w", fullName, err)
		}
	}

	return nil
}
