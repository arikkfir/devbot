package justest_test

import (
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
)

func TestNumericValueExtractor(t *testing.T) {
	type testCase struct {
		expectFailurePattern *string
		expectPanicPattern   *string
		actual               any
		expected             any
	}
	testCases := map[string]testCase{
		"string fails": {
			expectFailurePattern: lang.Ptr(`Unsupported actual value: 1`),
			actual:               "1",
		},
		"chan int": {
			actual:   lang.ChanOf(1),
			expected: 1,
		},
		"chan string fails": {
			expectFailurePattern: lang.Ptr(`Unsupported actual value: 1`),
			actual:               lang.ChanOf("1"),
		},
		"float32": {
			actual:   float32(1.1),
			expected: float32(1.1),
		},
		"float64": {
			actual:   1.1,
			expected: 1.1,
		},
		"func int": {
			actual:   func(t TT) any { return 1 },
			expected: 1,
		},
		"func string fails": {
			expectFailurePattern: lang.Ptr(`Unsupported actual value: 1`),
			actual:               func(t TT) any { return "1" },
		},
		"int": {
			actual:   1,
			expected: 1,
		},
		"int8": {
			actual:   int8(1),
			expected: int8(1),
		},
		"int16": {
			actual:   int16(1),
			expected: int16(1),
		},
		"int32": {
			actual:   int32(1),
			expected: int32(1),
		},
		"int64": {
			actual:   int64(1),
			expected: int64(1),
		},
		"pointer int": {
			actual:   lang.Ptr(1),
			expected: 1,
		},
		"pointer string fails": {
			expectFailurePattern: lang.Ptr(`Unsupported actual value: 1`),
			actual:               lang.Ptr("1"),
		},
		"uint": {
			actual:   1,
			expected: 1,
		},
		"uint8": {
			actual:   uint8(1),
			expected: uint8(1),
		},
		"uint16": {
			actual:   uint16(1),
			expected: uint16(1),
		},
		"uint32": {
			actual:   uint32(1),
			expected: uint32(1),
		},
		"uint64": {
			actual:   uint64(1),
			expected: uint64(1),
		},
	}
	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			mt := &MockT{Parent: NewTT(t)}
			if tc.expectFailurePattern != nil {
				if tc.expectPanicPattern != nil {
					t.Fatalf("Cannot expect both failure and panic")
				}
				defer expectFailure(t, mt, *tc.expectFailurePattern)
			}
			if tc.expectPanicPattern != nil {
				defer expectPanic(t, *tc.expectPanicPattern)
			}
			ve := NewNumericValueExtractor()
			v := ve.MustExtractValue(mt, tc.actual)
			if v != tc.expected {
				t.Fatalf("Expected '%v', got '%v'", tc.expected, v)
			}
		})
	}
}
