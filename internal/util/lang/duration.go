package lang

import (
	"fmt"
	"time"
)

func ParseDuration(minDuration time.Duration, value string) (time.Duration, error) {
	if duration, err := time.ParseDuration(value); err != nil {
		return 0, err
	} else if duration < minDuration {
		return 0, fmt.Errorf("refresh interval '%s' is too low (must not be less than %s)", value, minDuration)
	} else {
		return duration, nil
	}
}
