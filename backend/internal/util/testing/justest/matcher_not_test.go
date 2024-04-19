package justest_test

import (
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestNot(t *testing.T) {
	type Verifier func(t TT)
	type testCase struct {
		expectFailurePattern *string
		expectPanicPattern   *string
		actualsGenerator     func(t TT, tc *testCase) []any
		matcherGenerator     func(t TT, tc *testCase) (Matcher, Verifier)
		expectedResults      *[]any
	}
	testCases := map[string]testCase{
		"Failed matcher succeeds": {
			actualsGenerator: func(t TT, tc *testCase) []any { return []any{"foo-bar"} },
			matcherGenerator: func(t TT, tc *testCase) (Matcher, Verifier) {
				return func(t TT, actual ...any) []any { t.Fatalf("failure negated"); panic("unreachable") }, nil
			},
			expectedResults: lang.Ptr([]any{"foo-bar"}),
		},
		"Successful matcher fails": {
			expectFailurePattern: lang.Ptr[string]("Expected this matcher to fail, but it did not"),
			actualsGenerator:     func(t TT, tc *testCase) []any { return []any{"foo-bar"} },
			matcherGenerator: func(t TT, tc *testCase) (Matcher, Verifier) {
				return func(t TT, actual ...any) []any { return actual }, nil
			},
		},
		"Panicking matcher re-panics": {
			expectPanicPattern: lang.Ptr[string]("panic propagated"),
			actualsGenerator:   func(t TT, tc *testCase) []any { return []any{"foo-bar"} },
			matcherGenerator: func(t TT, tc *testCase) (Matcher, Verifier) {
				return func(t TT, actual ...any) []any { panic("panic propagated") }, nil
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			mt := NewMockT(NewTT(t))
			if tc.expectFailurePattern != nil {
				if tc.expectPanicPattern != nil {
					t.Fatalf("Invalid test - cannot specify both expected panic and expected failure")
				}
				defer expectFailure(t, mt, *tc.expectFailurePattern)
			} else if tc.expectPanicPattern != nil {
				defer expectPanic(t, *tc.expectPanicPattern)
			} else {
				defer expectNoFailure(t, mt)
			}
			matcher, verifier := tc.matcherGenerator(mt, &tc)
			actuals := tc.actualsGenerator(mt, &tc)
			result := For(mt).Expect(actuals...).Will(Not(matcher)).Result()
			if verifier != nil {
				verifier(mt)
			}
			if tc.expectedResults != nil {
				if !cmp.Equal(result, *tc.expectedResults) {
					t.Fatalf("expected result %+v, got %+v", *tc.expectedResults, result)
				}
			}
		})
	}
}
