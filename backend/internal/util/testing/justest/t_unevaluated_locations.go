package justest

import (
	"bytes"
	"fmt"
	"github.com/alecthomas/chroma/v2/quick"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	unevaluated = unevaluatedLocations{}
)

type unevaluatedLocation struct {
	function string
	file     string
	line     int
}

type unevaluatedLocations struct {
	values map[*testing.T]map[string]unevaluatedLocation
	mu     sync.RWMutex
}

//go:noinline
func (s *unevaluatedLocations) Add(t T, function, file string, line int) {
	GetHelper(t).Helper()
	s.mu.Lock()
	defer s.mu.Unlock()

	root := RootOf(t)
	if s.values == nil {
		s.values = make(map[*testing.T]map[string]unevaluatedLocation)
	}

	locations, locationsFound := s.values[root]
	if !locationsFound {
		locations = make(map[string]unevaluatedLocation, 10)
		s.values[root] = locations
		root.Cleanup(func() { s.print(root) })
	}

	key := fmt.Sprintf("%s:::%s:::%d", function, file, line)
	if _, ok := locations[key]; !ok {
		locations[key] = unevaluatedLocation{function: function, file: file, line: line}
	}
}

//go:noinline
func (s *unevaluatedLocations) print(t *testing.T) {
	GetHelper(t).Helper()
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.values == nil {
		t.Log("There should have been unevaluated locations but none were found! (for any T instance in this process)")
	} else if locations, ok := s.values[t]; !ok {
		t.Log("There should have been unevaluated locations but none were found! (for this T instance)")
	} else {
		for _, loc := range locations {
			source := "<could not read source>"
			if b, err := os.ReadFile(loc.file); err == nil {
				fileContents := string(b)
				lines := strings.Split(fileContents, "\n")
				if len(lines) > loc.line {
					source = strings.TrimSpace(lines[loc.line-1])
					output := bytes.Buffer{}
					if err := quick.Highlight(&output, source, "go", goSourceFormatter, goSourceStyle[displayMode]); err == nil {
						source = output.String()
					}
				}
			}
			_, _ = fmt.Fprintf(os.Stderr, "\t%s:%d: %s <-- Unevaluated expectation!\n", filepath.Base(loc.file), loc.line, source)
		}
		t.Fatalf("There were unevaluated test statements")
	}
}
