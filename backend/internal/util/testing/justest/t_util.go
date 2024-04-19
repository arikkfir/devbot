package justest

import (
	"fmt"
	"testing"
)

const (
	goSourceFormatter = "terminal256"
)

var (
	goSourceStyle = map[displayModeType]string{
		displayModeLight: "autumn",
		displayModeDark:  "catppuccin-mocha",
	}
)

var (
	ignoredStackTracePrefixes = []string{
		"testing.",
		"github.com/arikkfir/devbot/backend/internal/util/testing/justest/",
		"github.com/arikkfir/devbot/backend/internal/util/testing/justest.",
	}
)

//goland:noinspection GoUnusedExportedFunction
//go:noinline
func RootOf(t T) *testing.T {
	GetHelper(t).Helper()
	switch tt := t.(type) {
	case *tImpl:
		return RootOf(tt.t)
	case *inverseTT:
		return RootOf(tt.parent)
	case *eventuallyT:
		return RootOf(tt.parent)
	case *MockT:
		return RootOf(tt.Parent)
	case *testing.T:
		return tt
	default:
		panic(fmt.Sprintf("unrecognized TT type: %T", t))
	}
}
