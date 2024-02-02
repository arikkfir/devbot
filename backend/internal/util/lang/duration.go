package lang

import (
	"github.com/secureworks/errors"
	"time"
)

func ParseDuration(minSeconds int, value string) (time.Duration, error) {
	if duration, err := time.ParseDuration(value); err != nil {
		return 0, err
	} else if duration.Seconds() < float64(minSeconds) {
		return 0, errors.New("refresh interval '%s' is too low (must not be less than %ss)", value, minSeconds)
	} else {
		return duration, nil
	}
}
