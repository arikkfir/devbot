package justest

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type displayModeType string

const (
	displayModeLight displayModeType = "light"
	displayModeDark  displayModeType = "dark"

	appleScriptDarkModeQuery = `tell application "System Events" to tell appearance preferences to get dark mode`
)

var (
	displayMode = displayModeLight
)

func init() {
	cmd := exec.Command("osascript", "-e", appleScriptDarkModeQuery)
	if out, err := cmd.Output(); err != nil {
		displayMode = displayModeLight
		fmt.Printf("Error determining system's dark mode: %+v\n", err)
	} else if dark, err := strconv.ParseBool(strings.TrimSpace(string(out))); err != nil {
		displayMode = displayModeLight
		fmt.Printf("Error determining system's dark mode: %+v\n", err)
	} else if dark {
		displayMode = displayModeDark
	} else {
		displayMode = displayModeLight
	}
}
