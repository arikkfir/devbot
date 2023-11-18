package util

import (
	"regexp"
	"strings"
)

var (
	forbiddenSlugCharacters = regexp.MustCompile("[^a-zA-Z0-9]+")
)

//goland:noinspection GoUnusedExportedFunction
func Slugify(s string) string {
	forbiddenCharactersRemoved := forbiddenSlugCharacters.ReplaceAllString(s, "-")
	leadingTrailingDashesRemoved := strings.Trim(forbiddenCharactersRemoved, "-")
	lowerCased := strings.ToLower(leadingTrailingDashesRemoved)
	return lowerCased
}
