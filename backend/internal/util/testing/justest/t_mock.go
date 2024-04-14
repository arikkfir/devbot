package justest

import (
	"fmt"
	"time"
)

var (
	MockFatalPanic = fatalPanic{}
)

type MockT struct {
	Parent      TT
	FatalFormat string
	FatalArgs   []any
}

type fatalPanic struct{}

func (t *MockT) Cleanup(f func()) {
	GetHelper(t).Helper()
	t.Parent.Cleanup(f)
}

func (t *MockT) Fatalf(format string, args ...interface{}) {
	GetHelper(t).Helper()
	t.FatalFormat = format
	t.FatalArgs = args
	panic(MockFatalPanic)
}

func (t *MockT) Log(args ...any) {
	GetHelper(t).Helper()
	t.Parent.Log(args...)
}

func (t *MockT) Logf(format string, args ...any) {
	GetHelper(t).Helper()
	t.Parent.Logf(format, args...)
}

func (t *MockT) Deadline() (deadline time.Time, ok bool) {
	GetHelper(t).Helper()
	return t.Parent.Deadline()
}

func (t *MockT) Done() <-chan struct{} { GetHelper(t).Helper(); return t.Parent.Done() }

func (t *MockT) Err() error { GetHelper(t).Helper(); return t.Parent.Err() }

func (t *MockT) Value(key any) interface{} { GetHelper(t).Helper(); return getValueForT(t, key) }

func (t *MockT) GetHelper() Helper {
	if hp, ok := t.Parent.(HelperProvider); ok {
		return hp.GetHelper()
	} else if h, ok := t.Parent.(Helper); ok {
		return h
	} else {
		panic(fmt.Sprintf("could not obtain a HelperProvider instance from: %+v", t.Parent))
	}
}
