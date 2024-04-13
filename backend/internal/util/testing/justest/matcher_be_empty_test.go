package justest_test

import (
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"reflect"
	"regexp"
	"testing"
)

func TestBeEmpty(t *testing.T) {
	type testCase struct {
		expectFailurePattern *string
		actual               any
	}
	//goland:noinspection GoRedundantConversion
	testCases := map[reflect.Kind]map[string]testCase{
		reflect.Array: {
			"Empty succeeds":  {actual: [0]int{}},
			"Non empty fails": {actual: [3]int{1, 2, 3}, expectFailurePattern: lang.Ptr(regexp.QuoteMeta(`Expected '[1 2 3]' to be empty, but it is not (has a length of 3)`))},
		},
		reflect.Chan: {
			"Empty succeeds":  {actual: lang.ChanOf[int]()},
			"Non empty fails": {actual: lang.ChanOf[int](1, 2, 3), expectFailurePattern: lang.Ptr(`Expected '.+' to be empty, but it is not \(has a length of 3\)`)},
		},
		reflect.Map: {
			"Empty succeeds":  {actual: map[int]int{}},
			"Non empty fails": {actual: map[int]int{1: 1, 2: 2, 3: 3}, expectFailurePattern: lang.Ptr(regexp.QuoteMeta(`Expected 'map[1:1 2:2 3:3]' to be empty, but it is not (has a length of 3)`))},
		},
		reflect.Slice: {
			"Empty succeeds":  {actual: []int{}},
			"Non empty fails": {actual: []int{1, 2, 3}, expectFailurePattern: lang.Ptr(regexp.QuoteMeta(`Expected '[1 2 3]' to be empty, but it is not (has a length of 3)`))},
		},
		reflect.String: {
			"Empty succeeds":  {actual: ""},
			"Non empty fails": {actual: "abc", expectFailurePattern: lang.Ptr(regexp.QuoteMeta(`Expected 'abc' to be empty, but it is not (has a length of 3)`))},
		},
	}
	for kind, kindTestCases := range testCases {
		t.Run(kind.String(), func(t *testing.T) {
			t.Parallel()
			for name, tc := range kindTestCases {
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					mt := &MockT{Parent: NewTT(t)}
					if tc.expectFailurePattern != nil {
						defer expectFailure(t, mt, *tc.expectFailurePattern)
					} else {
						defer expectNoFailure(t, mt)
					}
					For(mt).Expect(tc.actual).Will(BeEmpty())
				})
			}
		})
	}
}
