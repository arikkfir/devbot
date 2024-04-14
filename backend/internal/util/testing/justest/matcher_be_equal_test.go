package justest_test

import (
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
)

func TestBeEqualTo(t *testing.T) {
	type testCase struct {
		expectFailurePattern *string
		actual, expected     any
	}
	testCases := map[string]testCase{
		"Equality succeeds": {actual: 1, expected: 1},
		"Difference fails":  {actual: 1, expected: 2, expectFailurePattern: lang.Ptr(`Expected & actual value differ:.*`)},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			mt := &MockT{Parent: NewTT(t)}
			if tc.expectFailurePattern != nil {
				defer expectFailure(t, mt, *tc.expectFailurePattern)
			} else {
				defer expectNoFailure(t, mt)
			}
			For(mt).Expect(tc.actual).Will(BeEqualTo(tc.expected)).OrFail()
		})
	}
}

func TestCompareTo(t *testing.T) {
	t.Run("Correct values passed", func(t *testing.T) {
		mt := &MockT{Parent: NewTT(t)}
		defer expectNoFailure(t, mt)
		comparator := func(t TT, expected, actual any) any {
			if expected != 2 {
				t.Fatalf("Incorrect 'expected' value provided - should be 1, got %+v", expected)
			}
			if actual != 1 {
				t.Fatalf("Incorrect 'actual' value provided - should be 2, got %+v", actual)
			}
			return actual
		}
		For(mt).Expect(1).Will(CompareTo(2).Using(comparator))
	})
	t.Run("Failure propagation", func(t *testing.T) {
		mt := &MockT{Parent: NewTT(t)}
		defer expectFailure(t, mt, `expected error`)
		comparator := func(t TT, expected, actual any) any {
			t.Fatalf("expected error")
			panic("unreachable")
		}
		For(mt).Expect(1).Will(CompareTo(2).Using(comparator)).OrFail()
	})
	t.Run("Success propagation", func(t *testing.T) {
		mt := &MockT{Parent: NewTT(t)}
		defer expectNoFailure(t, mt)
		called := false
		comparator := func(t TT, expected, actual any) any {
			called = true
			return actual
		}
		For(mt).Expect(1).Will(CompareTo(2).Using(comparator)).OrFail()
		if !called {
			t.Fatalf("Comparator was not called")
		}
	})
}
