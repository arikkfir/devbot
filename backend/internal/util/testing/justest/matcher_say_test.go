package justest_test

import (
	"bytes"
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"regexp"
	"testing"
)

func TestSay(t *testing.T) {
	type testCase struct {
		expectFailurePattern *string
		expectPanicPattern   *string
		actualString         string
		actualsGenerator     func(t TT, tc *testCase) []any
		stringRegexp         *string
		compiledRegexp       *regexp.Regexp
	}
	testCases := map[string]testCase{
		"*Buffer provides string actual": {
			actualString:     "abc",
			actualsGenerator: func(t TT, tc *testCase) []any { return []any{bytes.NewBuffer([]byte(tc.actualString))} },
			stringRegexp:     lang.Ptr("^abc$"),
		},
		"string provides string actual": {
			actualString:     "abc",
			actualsGenerator: func(t TT, tc *testCase) []any { return []any{tc.actualString} },
			stringRegexp:     lang.Ptr("^abc$"),
		},
		"regex string is used for matching": {
			actualString:     "abbbbc",
			actualsGenerator: func(t TT, tc *testCase) []any { return []any{tc.actualString} },
			stringRegexp:     lang.Ptr("^ab+c$"),
		},
		"compiled regex is used for matching": {
			actualString:     "abbbbc",
			actualsGenerator: func(t TT, tc *testCase) []any { return []any{tc.actualString} },
			compiledRegexp:   regexp.MustCompile("^ab+c$"),
		},
		"*string provides string actual": {
			actualString:     "abc",
			actualsGenerator: func(t TT, tc *testCase) []any { return []any{lang.Ptr(tc.actualString)} },
			stringRegexp:     lang.Ptr("^abc$"),
		},
		"*[]byte provides string actual": {
			actualString:     "abc",
			actualsGenerator: func(t TT, tc *testCase) []any { return []any{lang.Ptr([]byte(tc.actualString))} },
			stringRegexp:     lang.Ptr("^abc$"),
		},
		"[]byte provides string actual": {
			actualString:     "abc",
			actualsGenerator: func(t TT, tc *testCase) []any { return []any{[]byte(tc.actualString)} },
			stringRegexp:     lang.Ptr("^abc$"),
		},
		"Non-bytes slice fails": {
			expectFailurePattern: lang.Ptr(`Unsupported type '\[\]int' for Say matcher: \[1 2 3\]`),
			actualString:         "abc",
			actualsGenerator:     func(t TT, tc *testCase) []any { return []any{[]int{1, 2, 3}} },
			stringRegexp:         lang.Ptr("^abc$"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			mt := &MockT{Parent: NewTT(t)}
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
			var results []any
			if tc.stringRegexp != nil {
				results = For(mt).Expect(actuals...).Will(Say(*tc.stringRegexp)).Result()
			} else if tc.compiledRegexp != nil {
				results = For(mt).Expect(actuals...).Will(Say(tc.compiledRegexp)).Result()
			}
			if len(results) != 1 {
				t.Fatalf("expected single result, got %+v", results)
			} else if results[0].(string) != tc.actualString {
				t.Fatalf("expected result %+v, got %+v", tc.actualString, results)
			}
		})
	}
}
