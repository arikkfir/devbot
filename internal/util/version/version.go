package version

import (
	"fmt"
	"golang.org/x/mod/semver"
)

// Version represents the version of the controller.
var Version string

func init() {
	if !semver.IsValid(Version) {
		panic(fmt.Errorf("invalid version injected: %s", Version))
	}
}
