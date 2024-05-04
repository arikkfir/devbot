package configuration

import (
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"strings"
)

func ApplyBoolEnvironmentVariableTo(target *bool, envVarName string) {
	if stringValue, found := os.LookupEnv(envVarName); found {
		if boolValue, err := strconv.ParseBool(stringValue); err != nil {
			log.Fatal().Err(err).Msgf("Failed to parse %s environment variable", envVarName)
		} else {
			*target = boolValue
		}
	}
}

func ApplyStringEnvironmentVariableTo(target *string, envVarName string) {
	if value, found := os.LookupEnv(envVarName); found {
		*target = value
	}
}

func ApplyIntEnvironmentVariableTo(target *int, envVarName string) {
	if stringValue, found := os.LookupEnv(envVarName); found {
		if intValue, err := strconv.ParseInt(stringValue, 10, 0); err != nil {
			log.Fatal().Err(err).Msgf("Failed to parse %s environment variable", envVarName)
		} else {
			*target = int(intValue)
		}
	}
}

func FlagNameToEnvironmentVariable(name string) string {
	return strings.ReplaceAll(strings.ToUpper(name), "-", "_")
}
