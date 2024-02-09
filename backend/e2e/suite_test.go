package e2e_test

import (
	"github.com/onsi/ginkgo/v2/types"
	"github.com/secureworks/errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func intEnv(name string, defaultValue int) int {
	if v, found := os.LookupEnv(name); found {
		if i, err := strconv.Atoi(v); err != nil {
			panic(errors.New("%s must be an integer, but was '%s'", name, v))
		} else {
			return i
		}
	}
	return defaultValue
}

func int64Env(name string, defaultValue int64) int64 {
	if v, found := os.LookupEnv(name); found {
		if i, err := strconv.Atoi(v); err != nil {
			panic(errors.New("%s must be an integer, but was '%s'", name, v))
		} else {
			return int64(i)
		}
	}
	return defaultValue
}

func boolEnv(name string, defaultValue bool) bool {
	if v, found := os.LookupEnv(name); found {
		if b, err := strconv.ParseBool(v); err != nil {
			panic(errors.New("%s must be a boolean, but was '%s'", name, v))
		} else {
			return b
		}
	}
	return defaultValue
}

func stringArrayEnv(name, separator string, defaultValue []string) []string {
	if v, found := os.LookupEnv(name); found {
		return strings.Split(v, separator)
	}
	return defaultValue
}

func stringEnv(name string, defaultValue string) string {
	if v, found := os.LookupEnv(name); found {
		return v
	}
	return defaultValue
}

func durationEnv(name string, defaultValue time.Duration) time.Duration {
	if v, found := os.LookupEnv(name); found {
		if d, err := time.ParseDuration(v); err != nil {
			panic(errors.New("%s must be a duration, but was '%s'", name, v))
		} else {
			return d
		}
	}
	return defaultValue
}

func TestEndToEnd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(
		t,
		"End to End",
		types.SuiteConfig{
			DryRun:               boolEnv("GINKGO_DRY_RUN", false),
			FailFast:             boolEnv("GINKGO_FAIL_FAST", true),
			FailOnPending:        boolEnv("GINKGO_FAIL_ON_PENDING", false),
			FlakeAttempts:        intEnv("GINKGO_FLAKE_ATTEMPTS", 0),
			FocusFiles:           stringArrayEnv("GINKGO_FOCUS_FILES", ",", nil),
			FocusStrings:         stringArrayEnv("GINKGO_FOCUS_STRINGS", ",", nil),
			GracePeriod:          durationEnv("GINKGO_TIMEOUT", 30*time.Second),
			LabelFilter:          stringEnv("GINKGO_LABEL_FILTER", ""),
			MustPassRepeatedly:   intEnv("GINKGO_MUST_PASS_REPEATEDLY", 0),
			ParallelProcess:      1,
			ParallelTotal:        1,
			PollProgressAfter:    durationEnv("GINKGO_POLL_PROGRESS_AFTER", 0),
			PollProgressInterval: durationEnv("GINKGO_POLL_PROGRESS_INTERVAL", 0),
			RandomSeed:           int64Env("GINKGO_RANDOM_SEED", 0),
			RandomizeAllSpecs:    boolEnv("GINKGO_RANDOMIZE_ALL_SPECS", false),
			SkipFiles:            stringArrayEnv("GINKGO_SKIP_FILES", ",", nil),
			SkipStrings:          stringArrayEnv("GINKGO_SKIP_STRINGS", ",", nil),
			Timeout:              durationEnv("GINKGO_TIMEOUT", 0),
		},
		types.ReporterConfig{
			FullTrace:      boolEnv("GINKGO_FULL_TRACE", false),
			JSONReport:     stringEnv("GINKGO_JSON_REPORT", ""),
			JUnitReport:    stringEnv("GINKGO_JUNIT_REPORT", ""),
			NoColor:        boolEnv("GINKGO_NO_COLOR", false),
			ShowNodeEvents: boolEnv("GINKGO_SHOW_NODE_EVENTS", false),
			Succinct:       boolEnv("GINKGO_SUCCINCT", false),
			TeamcityReport: stringEnv("GINKGO_TEAMCITY_REPORT", ""),
			Verbose:        boolEnv("GINKGO_VERBOSE", false),
			VeryVerbose:    boolEnv("GINKGO_VERY_VERBOSE", false),
		},
	)
}
