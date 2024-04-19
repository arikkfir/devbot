package justest_test

import (
	"fmt"
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestSucceed(t *testing.T) {
	type testCase struct {
		expectFailurePattern *string
		expectPanicPattern   *string
		actualsGenerator     func(t TT, tc *testCase) []any
		expectedResults      *[]any
	}
	testCases := map[string]testCase{
		"Succeeds if no actuals": {
			actualsGenerator: func(t TT, tc *testCase) []any { return []any{} },
			expectedResults:  lang.Ptr([]any{}),
		},
		"Succeeds if last actual is nil and returns without last item": {
			actualsGenerator: func(t TT, tc *testCase) []any { return []any{1, 2, 3} },
			expectedResults:  lang.Ptr([]any{1, 2}),
		},
		"Fails if last actual is an error": {
			expectFailurePattern: lang.Ptr[string]("Error occurred: expected error"),
			actualsGenerator:     func(t TT, tc *testCase) []any { return []any{"abc", fmt.Errorf("expected error")} },
		},
		"Succeeds if last actual is not an error and returns without last item": {
			actualsGenerator: func(t TT, tc *testCase) []any { return []any{"abc", "def"} },
			expectedResults:  lang.Ptr([]any{"abc"}),
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
			actuals := tc.actualsGenerator(mt, &tc)
			result := For(mt).Expect(actuals...).Will(Succeed()).Result()
			if tc.expectedResults != nil {
				if !cmp.Equal(result, *tc.expectedResults) {
					t.Fatalf("expected result %+v, got %+v", *tc.expectedResults, result)
				}
			}
		})
	}
}
