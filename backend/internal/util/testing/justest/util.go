package justest

import (
	"fmt"
	"regexp"
)

type FormatAndArgs struct {
	Format *string
	Args   []any
}

//go:noinline
func (f FormatAndArgs) String() string {
	if f.Format != nil {
		return fmt.Sprintf(*f.Format, f.Args...)
	} else {
		return fmt.Sprint(f.Args...)
	}
}

type TestOutcomeExpectation string

const (
	ExpectFailure TestOutcomeExpectation = "expect failure"
	ExpectPanic   TestOutcomeExpectation = "expect panic"
	ExpectSuccess TestOutcomeExpectation = "expect success"
)

//go:noinline
func VerifyTestOutcome(t T, expectedOutcome TestOutcomeExpectation, pattern string) {
	GetHelper(t).Helper()
	if t.Failed() {
		// If the given T has already failed, there's no point verifying the mock T (
		// which would be a potential panic recovered)
		return
	}

	switch expectedOutcome {
	case ExpectFailure:
		if r := recover(); r == nil {
			t.Fatalf("Expected test failure did not happen")
		} else if mt, ok := r.(*MockT); !ok {
			t.Fatalf("Unexpected panic '%+v' happened instead of an expected test failure", mt)
		} else {
			for _, f := range mt.Failures {
				if actualMsg := f.String(); !regexp.MustCompile(pattern).MatchString(actualMsg) {
					t.Fatalf("Expected test failure matching '%s', but got: %s", pattern, actualMsg)
				}
			}
		}
	case ExpectPanic:
		if r := recover(); r == nil {
			t.Fatalf("Expected panic did not happen")
		} else if msg := fmt.Sprintf("%+v", r); !regexp.MustCompile(pattern).MatchString(msg) {
			t.Fatalf("Expected panic matching '%s', but got: %s", pattern, msg)
		}
	case ExpectSuccess:
		if r := recover(); r == nil {
			// ok; no-op
		} else if mt, ok := r.(*MockT); !ok {
			t.Fatalf("Unexpected panic '%+v' happened", mt)
		} else {
			msg := ""
			for _, f := range mt.Failures {
				msg = fmt.Sprintf("%s\n%s", msg, f.String())
			}
			t.Fatalf("Test failure(s) happened when no test failures were expected:%s", msg)
		}
	}
}
