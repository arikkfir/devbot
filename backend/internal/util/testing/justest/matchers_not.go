package justest

import (
	"github.com/secureworks/errors"
)

type inverseT struct {
	parent  T
	failure *FormatAndArgs
}

//go:noinline
func (t *inverseT) Cleanup(f func()) { GetHelper(t).Helper(); t.parent.Cleanup(f) }

//go:noinline
func (t *inverseT) Failed() bool {
	GetHelper(t.parent).Helper()
	return t.failure != nil
}

//go:noinline
func (t *inverseT) Fatalf(format string, args ...any) {
	GetHelper(t.parent).Helper()
	t.failure = &FormatAndArgs{&format, args}
	panic(t)
}

//go:noinline
func (t *inverseT) Log(args ...any) {
	GetHelper(t.parent).Helper()
	t.parent.Log(args...)
}

//go:noinline
func (t *inverseT) Logf(format string, args ...any) {
	GetHelper(t.parent).Helper()
	t.parent.Logf(format, args...)
}

//go:noinline
func (t *inverseT) GetParent() T {
	return t.parent
}

//go:noinline
func Not(m Matcher) Matcher {
	return MatcherFunc(func(t T, actuals ...any) {
		GetHelper(t).Helper()

		tt := &inverseT{parent: t}

		defer func() {
			GetHelper(t).Helper()
			if r := recover(); r != nil {
				if r == tt {
					// Matcher failed - good!
				} else {
					// Unexpected panic - bubble it up
					panic(errors.New("unexpected panic: %+v", r))
				}
			} else {
				t.Fatalf("Expected this matcher to fail, but it did not")
			}
		}()

		m.Assert(tt, actuals...)
	})
}