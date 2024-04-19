package justest

import (
	"fmt"
	"github.com/secureworks/errors"
	"time"
)

type inverseTT struct {
	parent    TT
	fatalMsg  *string
	fatalArgs *[]any
	cleanup   []func()
}

//go:noinline
func (t *inverseTT) PerformCleanups() {
	GetHelper(t).Helper()
	cleanups := t.cleanup
	t.cleanup = nil
	for i := len(cleanups) - 1; i >= 0; i-- {
		cleanups[i]()
	}
}

//go:noinline
func (t *inverseTT) Cleanup(f func()) { GetHelper(t).Helper(); t.cleanup = append(t.cleanup, f) }

//go:noinline
func (t *inverseTT) Fatalf(format string, args ...any) {
	GetHelper(t.parent).Helper()
	f := format
	a := args
	t.fatalMsg = &f
	t.fatalArgs = &a
	panic(t)
}

//go:noinline
func (t *inverseTT) Log(args ...any) {
	GetHelper(t.parent).Helper()
	t.parent.Log(args...)
}

//go:noinline
func (t *inverseTT) Logf(format string, args ...any) {
	GetHelper(t.parent).Helper()
	t.parent.Logf(format, args...)
}

//go:noinline
func (t *inverseTT) GetHelper() Helper {
	if hp, ok := t.parent.(HelperProvider); ok {
		return hp.GetHelper()
	} else if h, ok := t.parent.(Helper); ok {
		return h
	} else {
		panic(fmt.Sprintf("could not obtain a HelperProvider instance from: %+v", t.parent))
	}
}

//go:noinline
func (t *inverseTT) Deadline() (deadline time.Time, ok bool) {
	GetHelper(t).Helper()
	return t.parent.Deadline()
}

//go:noinline
func (t *inverseTT) Done() <-chan struct{} { GetHelper(t).Helper(); return t.parent.Done() }

//go:noinline
func (t *inverseTT) Err() error { GetHelper(t).Helper(); return t.parent.Err() }

//go:noinline
func (t *inverseTT) Value(key any) interface{} { GetHelper(t).Helper(); return getValueForT(t, key) }

//go:noinline
func Not(m Matcher) Matcher {
	return func(t TT, actuals ...any) (results []any) {
		GetHelper(t).Helper()

		tt := &inverseTT{parent: t}
		t.Cleanup(tt.PerformCleanups)

		defer func() {
			GetHelper(t).Helper()
			if r := recover(); r != nil {
				if r == tt && tt.fatalMsg != nil {
					// Matcher failed - good!
				} else {
					// Unexpected panic - bubble it up
					panic(errors.New("unexpected panic! recovered: %+v", r))
				}
			} else {
				t.Fatalf("Expected this matcher to fail, but it did not")
			}
		}()

		results = actuals
		_ = m(tt, actuals...)
		return
	}
}
