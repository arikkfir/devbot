package justest_test

import (
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"github.com/google/go-cmp/cmp"
	"testing"
	"time"
)

func TestEventually(t *testing.T) {
	type Verifier func(t TT)
	type testCase struct {
		expectFailurePattern *string
		expectPanicPattern   *string
		actualsGenerator     func(t TT, tc *testCase) []any
		matcherGenerator     func(t TT, tc *testCase) (Matcher, Verifier)
		timeout              time.Duration
		interval             time.Duration
		expectedResults      *[]any
	}
	testCases := map[string]testCase{
		"Expectation (matcher) is required": {
			expectPanicPattern: lang.Ptr(`expectation cannot be nil`),
			actualsGenerator:   func(t TT, tc *testCase) []any { return []any{"foo-bar"} },
			matcherGenerator:   func(t TT, tc *testCase) (Matcher, Verifier) { return nil, nil },
			timeout:            100 * time.Millisecond,
			interval:           10 * time.Millisecond,
		},
		"Interval cannot be zero": {
			expectPanicPattern: lang.Ptr(`interval cannot be 0`),
			actualsGenerator:   func(t TT, tc *testCase) []any { return []any{"foo-bar"} },
			matcherGenerator: func(t TT, tc *testCase) (Matcher, Verifier) {
				return func(t TT, actual ...any) []any { return actual }, nil
			},
			timeout:  100 * time.Millisecond,
			interval: 0,
		},
		"Interval cannot be greater than timeout": {
			expectPanicPattern: lang.Ptr(`interval cannot be greater than timeout`),
			actualsGenerator:   func(t TT, tc *testCase) []any { return []any{"foo-bar"} },
			matcherGenerator: func(t TT, tc *testCase) (Matcher, Verifier) {
				return func(t TT, actual ...any) []any { return actual }, nil
			},
			timeout:  100 * time.Millisecond,
			interval: 1 * time.Second,
		},
		"Cleanup is local": {
			actualsGenerator: func(t TT, tc *testCase) []any { return []any{"foo-bar"} },
			matcherGenerator: func(t TT, tc *testCase) (Matcher, Verifier) {
				cleanups := make(chan any, 100)
				matcher := func(t TT, actual ...any) []any {
					t.Cleanup(func() { cleanups <- t.(RetryingTT).Trial() })
					return actual
				}
				verifier := func(t TT) {
					close(cleanups)
					expected := []any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
					for v := range cleanups {
						if v != expected[0] {
							t.Fatalf("expected %v, got %v", expected[0], v)
						} else {
							expected = expected[1:]
						}
					}
				}
				return matcher, verifier
			},
			timeout:  500 * time.Millisecond,
			interval: 100 * time.Millisecond,
		},
		"Fails on timeout": {
			expectFailurePattern: lang.Ptr(`Timed out after [0-9]+(\.[0-9]+)?s waiting for expectation to be met:\n\t.+ <-- failure`),
			actualsGenerator:     func(t TT, tc *testCase) []any { return []any{"foo-bar"} },
			matcherGenerator: func(t TT, tc *testCase) (Matcher, Verifier) {
				var firstTrial *time.Time
				return func(t TT, actual ...any) []any {
						if t.(RetryingTT).Trial() == 1 {
							firstTrial = lang.Ptr(time.Now())
						}
						t.Fatalf("failure")
						panic("unreachable")
					}, func(t TT) {
						if firstTrial == nil {
							t.Fatalf("first trial was not recorded")
						}
						timeSinceFirstTrial := time.Since(*firstTrial)
						if timeSinceFirstTrial < tc.timeout {
							t.Fatalf("Eventually failed after %s which is too early", timeSinceFirstTrial)
						}
					}
			},
			timeout:  5 * time.Second,
			interval: 100 * time.Millisecond,
		},
		"Unexpected panics are propagated": {
			expectPanicPattern: lang.Ptr(`unexpected panic! recovered: an expected panic`),
			actualsGenerator:   func(t TT, tc *testCase) []any { return []any{"foo-bar"} },
			matcherGenerator: func(t TT, tc *testCase) (Matcher, Verifier) {
				return func(t TT, actual ...any) []any { panic("an expected panic") }, nil
			},
			timeout:  1 * time.Second,
			interval: 100 * time.Millisecond,
		},
		"Correct result returned": {
			actualsGenerator: func(t TT, tc *testCase) []any { return []any{"foo-bar"} },
			matcherGenerator: func(t TT, tc *testCase) (Matcher, Verifier) {
				return func(t TT, actual ...any) []any { return actual }, nil
			},
			timeout:         1 * time.Second,
			interval:        100 * time.Millisecond,
			expectedResults: lang.Ptr([]any{"foo-bar"}),
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
			result := For(mt).Expect(actuals...).Will(Eventually(matcher).Within(tc.timeout).ProbingEvery(tc.interval)).Result()
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
