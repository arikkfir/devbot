package justest_test

import (
	"fmt"
	"github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
)

var (
	mockFatalPanic = fatalPanic{}
)

type MockT struct {
	parent      justest.JustT
	fatalFormat string
	fatalArgs   []any
}

type fatalPanic struct{}

func (t *MockT) Cleanup(f func()) {
	justest.GetHelper(t).Helper()
	t.parent.Cleanup(f)
}

func (t *MockT) Fatalf(format string, args ...interface{}) {
	justest.GetHelper(t).Helper()
	t.fatalFormat = format
	t.fatalArgs = args
	panic(mockFatalPanic)
}

func (t *MockT) GetHelper() justest.Helper {
	if hp, ok := t.parent.(justest.HelperProvider); ok {
		return hp.GetHelper()
	} else if h, ok := t.parent.(justest.Helper); ok {
		return h
	} else {
		panic(fmt.Sprintf("could not obtain a HelperProvider instance from the given JustT instance: %+v", t.parent))
	}
}

func (t *MockT) Log(args ...any) {
	justest.GetHelper(t).Helper()
	t.parent.Log(args...)
}

func (t *MockT) Logf(format string, args ...any) {
	justest.GetHelper(t).Helper()
	t.parent.Logf(format, args...)
}

func verifyTestCaseError(t *testing.T, mt *MockT, wantErr bool) {
	justest.GetHelper(t).Helper()
	if wantErr {
		if r := recover(); r == nil {
			t.Error("expected assertion to call 't.Fatalf(...)', but it did not")
		} else if r != mockFatalPanic {
			panic(r)
		}
	} else {
		if r := recover(); r != nil {
			if r != mockFatalPanic {
				panic(r)
			} else {
				t.Errorf("expected assertion not to fail, but it called t.Fatalf(%s, %v)", mt.fatalFormat, mt.fatalArgs)
			}
		}
	}
}
