package lang

import (
	"errors"
)

func IgnoreErrorOfType(err, ignored error) error {
	if errors.Is(err, ignored) {
		return nil
	} else {
		return err
	}
}
