package justest

import (
	"fmt"
	"github.com/secureworks/errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RetryingTT interface {
	TT
	Trial() int
}

type eventuallyT struct {
	trial   int
	parent  TT
	cleanup []func()
	error   *struct {
		msg  string
		args []any
	}
}

//go:noinline
func (t *eventuallyT) Trial() int { GetHelper(t).Helper(); return t.trial }

//go:noinline
func (t *eventuallyT) Cleanup(f func()) { GetHelper(t).Helper(); t.cleanup = append(t.cleanup, f) }

//go:noinline
func (t *eventuallyT) Fatalf(format string, args ...any) {
	GetHelper(t).Helper()

	for _, frame := range errors.CallStackAt(2) {
		function, file, line := frame.Location()

		startsWithAnIgnoredPrefix := false
		for _, prefix := range ignoredStackTracePrefixes {
			if strings.HasPrefix(function, prefix) {
				startsWithAnIgnoredPrefix = true
				break
			}
		}

		if !startsWithAnIgnoredPrefix {
			location := fmt.Sprintf("%s:%d", filepath.Base(file), line)

			// Try to read the actual statement that fails
			if bytes, err := os.ReadFile(file); err == nil {
				fileContents := string(bytes)
				lines := strings.Split(fileContents, "\n")
				if len(lines) > line {
					lineContents := strings.TrimSpace(lines[line-1])
					location = location + ": " + lineContents
				}
			}

			format = format + "\n%s"
			args = append(args, location)
			break
		}
	}
	t.error = &struct {
		msg  string
		args []any
	}{msg: format, args: args}
	panic(t)
}

//go:noinline
func (t *eventuallyT) Log(args ...any) { GetHelper(t).Helper(); t.parent.Log(args...) }

//go:noinline
func (t *eventuallyT) Logf(format string, args ...any) {
	GetHelper(t).Helper()
	t.parent.Logf(format, args...)
}

//go:noinline
func (t *eventuallyT) Deadline() (deadline time.Time, ok bool) {
	GetHelper(t).Helper()
	return t.parent.Deadline()
}

//go:noinline
func (t *eventuallyT) Done() <-chan struct{} { GetHelper(t).Helper(); return t.parent.Done() }

//go:noinline
func (t *eventuallyT) Err() error { GetHelper(t).Helper(); return t.parent.Err() }

//go:noinline
func (t *eventuallyT) Value(key any) interface{} {
	GetHelper(t).Helper()
	return getValueForT(t, key)
}

//go:noinline
func (t *eventuallyT) GetHelper() Helper {
	if hp, ok := t.parent.(HelperProvider); ok {
		return hp.GetHelper()
	} else if h, ok := t.parent.(Helper); ok {
		return h
	} else {
		panic(fmt.Sprintf("could not obtain a HelperProvider instance from the given JustT instance: %+v", t.parent))
	}
}
