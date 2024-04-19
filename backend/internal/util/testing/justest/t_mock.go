package justest

import (
	"fmt"
	"time"
)

var (
	MockFatalPanic = fatalPanic{}
)

//go:noinline
func NewMockT(t TT) *MockT {
	tt := &MockT{Parent: t}
	t.Cleanup(tt.PerformCleanups)
	return tt
}

type MockT struct {
	Parent      TT
	FatalFormat string
	FatalArgs   []any
	cleanup     []func()
}

type fatalPanic struct{}

//go:noinline
func (t *MockT) PerformCleanups() {
	GetHelper(t).Helper()
	cleanups := t.cleanup
	t.cleanup = nil
	for i := len(cleanups) - 1; i >= 0; i-- {
		cleanups[i]()
	}
}

//go:noinline
func (t *MockT) Cleanup(f func()) { GetHelper(t).Helper(); t.cleanup = append(t.cleanup, f) }

//go:noinline
func (t *MockT) Fatalf(format string, args ...interface{}) {
	GetHelper(t).Helper()
	t.FatalFormat = format
	t.FatalArgs = args
	panic(MockFatalPanic)
}

//go:noinline
func (t *MockT) Log(args ...any) {
	GetHelper(t).Helper()
	t.Parent.Log(args...)
}

//go:noinline
func (t *MockT) Logf(format string, args ...any) {
	GetHelper(t).Helper()
	t.Parent.Logf(format, args...)
}

//go:noinline
func (t *MockT) Deadline() (deadline time.Time, ok bool) {
	GetHelper(t).Helper()
	return t.Parent.Deadline()
}

//go:noinline
func (t *MockT) Done() <-chan struct{} { GetHelper(t).Helper(); return t.Parent.Done() }

//go:noinline
func (t *MockT) Err() error { GetHelper(t).Helper(); return t.Parent.Err() }

//go:noinline
func (t *MockT) Value(key any) interface{} { GetHelper(t).Helper(); return getValueForT(t, key) }

//go:noinline
func (t *MockT) GetHelper() Helper {
	if hp, ok := t.Parent.(HelperProvider); ok {
		return hp.GetHelper()
	} else if h, ok := t.Parent.(Helper); ok {
		return h
	} else {
		panic(fmt.Sprintf("could not obtain a HelperProvider instance from: %+v", t.Parent))
	}
}
