package justest_test

import (
	"fmt"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
	"time"
)

func TestEventually(t *testing.T) {
	type testCase struct {
		name                string
		positiveExpectation bool
		getExpectedValues   func(t *testing.T, tc *testCase) []any
		timeout             time.Duration
		interval            time.Duration
		matcher             Matcher
		wantErr             bool
	}
	cases := []testCase{
		{
			name:                "FuncReturningErrorWillSucceedWhenReturningNull",
			positiveExpectation: true,
			getExpectedValues: func(t *testing.T, tc *testCase) []any {
				var invocations = 0
				return []any{
					func(t JustT) error {
						invocations++
						if invocations >= 15 {
							return nil
						}
						return fmt.Errorf("failed")
					},
				}
			},
			timeout:  1 * time.Second,
			interval: 10 * time.Millisecond,
			matcher:  Succeed(),
			wantErr:  false,
		},
		{
			name:                "FuncReturningNilWillNotSucceedWhenReturningError",
			positiveExpectation: false,
			getExpectedValues: func(t *testing.T, tc *testCase) []any {
				return []any{
					func(t JustT) error {
						return fmt.Errorf("failed")
					},
				}
			},
			timeout:  1 * time.Second,
			interval: 10 * time.Millisecond,
			matcher:  Succeed(),
			wantErr:  false,
		},
		{
			name:                "FuncReturningIntWillSucceedWhenReturning15",
			positiveExpectation: true,
			getExpectedValues: func(t *testing.T, tc *testCase) []any {
				var invocations = 0
				return []any{
					func() int {
						invocations++
						return invocations
					},
				}
			},
			timeout:  1 * time.Second,
			interval: 10 * time.Millisecond,
			matcher:  BeEqualTo(15),
			wantErr:  false,
		},
		{
			name:                "FuncReturningIntWillNotEqual150AndFail",
			positiveExpectation: true,
			getExpectedValues: func(t *testing.T, tc *testCase) []any {
				var invocations = 0
				return []any{
					func() int {
						invocations++
						return invocations
					},
				}
			},
			timeout:  1 * time.Second,
			interval: 10 * time.Millisecond,
			matcher:  BeEqualTo(2000),
			wantErr:  true,
		},
		{
			name:                "FuncWithoutReturnValueWithFailedAssertionWillFail",
			positiveExpectation: true,
			getExpectedValues: func(t *testing.T, tc *testCase) []any {
				var invocations = 0
				return []any{
					func(t JustT) {
						invocations++
						For(t).Expect(invocations).Will(BeEqualTo(1500))
					},
				}
			},
			timeout:  1 * time.Second,
			interval: 10 * time.Millisecond,
			matcher:  Succeed(),
			wantErr:  true,
		},
		{
			name:                "FuncReturningNilErrorButWithFailedAssertionWillFail",
			positiveExpectation: true,
			getExpectedValues: func(t *testing.T, tc *testCase) []any {
				var invocations = 0
				return []any{
					func(t JustT) error {
						invocations++
						For(t).Expect(invocations).Will(BeEqualTo(1500))
						return nil
					},
				}
			},
			timeout:  1 * time.Second,
			interval: 10 * time.Millisecond,
			matcher:  Succeed(),
			wantErr:  true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mt := &MockT{parent: t}
			defer verifyTestCaseError(t, mt, tc.wantErr)
			if tc.positiveExpectation {
				For(mt).Expect(tc.getExpectedValues(t, &tc)...).Will(Eventually(tc.matcher).Within(tc.timeout).ProbingEvery(tc.interval))
			} else {
				For(mt).Expect(tc.getExpectedValues(t, &tc)...).WillNot(Eventually(tc.matcher).Within(tc.timeout).ProbingEvery(tc.interval))
			}
		})
	}
}
