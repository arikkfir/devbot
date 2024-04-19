package justest_test

import (
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestEntrypoint(t *testing.T) {
	type testCase struct {
		expectFailurePattern *string
		actuals              []any
		matcher              func(*testCase) Matcher
		calledMatcherActual  *any
	}
	testCases := map[string]testCase{
		"Expected values provided as-is to matcher": {
			actuals: []any{1, 2, 3},
			matcher: func(tc *testCase) Matcher {
				return func(t TT, actuals ...any) []any {
					if !cmp.Equal(actuals, tc.actuals) {
						t.Fatalf("Expected matcher to be called with actuals '%+v', but got '%+v'", tc.actuals, actuals)
					}
					return actuals
				}
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			mt := NewMockT(NewTT(t))
			if tc.expectFailurePattern != nil {
				defer expectFailure(t, mt, *tc.expectFailurePattern)
			} else {
				defer expectNoFailure(t, mt)
			}
			For(mt).Expect(tc.actuals...).Will(tc.matcher(&tc)).OrFail()
		})
	}

	t.Run("Values added & retrieved correctly for T", func(t *testing.T) {

		For(t).AddValue("k1", "v1")
		vForT := For(t).Value("k1")
		if vForT != "v1" {
			t.Fatalf("Expected value for 'testing.T' key 'k1' to be 'v1', but got '%v'", vForT)
		}

		tt := NewTT(t)
		vForTT := For(tt).Value("k1")
		if vForTT != nil {
			t.Fatalf("Expected value for 'tImpl(testing.T)' key 'k1' to be nil, but got '%v'", vForTT)
		}

	})

	t.Run("Explained", func(t *testing.T) {
		matcher := func(t TT, actuals ...any) []any {
			GetHelper(t).Helper()
			t.Fatalf("Failed!")
			panic("unreachable")
		}

		mt := NewMockT(NewTT(t))
		defer expectFailure(t, mt, `^Failed! \(this is just a test\)$`)
		For(mt).Expect(1).Will(matcher).Because("this is just a %s", "test").OrFail()
	})
}
