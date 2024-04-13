package justest_test

import (
	"fmt"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"regexp"
	"testing"
)

func expectFailure(t *testing.T, mt *MockT, msgPattern string) {
	r := recover()
	GetHelper(t).Helper()
	if r == nil {
		t.Fatal("Expected test failure did not happen")
	} else if r != MockFatalPanic {
		t.Fatalf("Unexpected panic '%+v' happened instead of an expected test failure matching '%s'", r, msgPattern)
	} else if msg := fmt.Sprintf(mt.FatalFormat, mt.FatalArgs...); !regexp.MustCompile(msgPattern).MatchString(msg) {
		t.Fatalf("Expected test failure matching '%s', but got: %s", msgPattern, msg)
	}
}

func expectNoFailure(t *testing.T, mt *MockT) {
	GetHelper(t).Helper()
	if r := recover(); r == nil {
		// ok; no-op
	} else if r != MockFatalPanic {
		t.Fatalf("Unexpected panic '%+v' happened", r)
	} else {
		t.Fatalf("Test failure happened when no test failure was expected: %s", fmt.Sprintf(mt.FatalFormat, mt.FatalArgs...))
	}
}

func expectPanic(t *testing.T, msgPattern string) {
	r := recover()
	GetHelper(t).Helper()
	if r == nil {
		t.Fatal("Expected panic did not happen")
	} else if msg := fmt.Sprintf("%+v", r); !regexp.MustCompile(msgPattern).MatchString(msg) {
		t.Fatalf("Expected panic matching '%s', but got: %s", msgPattern, msg)
	}
}

func verifyTestCaseError(t *testing.T, mt *MockT, wantErr bool) {
	GetHelper(t).Helper()
	if wantErr {
		if r := recover(); r == nil {
			t.Error("expected assertion to call 't.Fatalf(...)', but it did not")
		} else if r != MockFatalPanic {
			panic(r)
		}
	} else {
		if r := recover(); r != nil {
			if r != MockFatalPanic {
				panic(r)
			} else {
				t.Errorf("expected assertion not to fail, but it called t.Fatalf(%s, %v)", mt.FatalFormat, mt.FatalArgs)
			}
		}
	}
}
