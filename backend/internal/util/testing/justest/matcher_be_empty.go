package justest

import (
	"context"
	"reflect"
)

var (
	emptyValueExtractor ValueExtractor
)

func init() {
	emptyValueExtractor = NewValueExtractor(ExtractorUnsupported)
	emptyValueExtractor[reflect.Array] = ExtractorValueLen
	emptyValueExtractor[reflect.Chan] = ExtractorValueLen
	emptyValueExtractor[reflect.Map] = ExtractorValueLen
	emptyValueExtractor[reflect.Pointer] = NewPointerExtractor(emptyValueExtractor, true)
	emptyValueExtractor[reflect.Slice] = ExtractorValueLen
	emptyValueExtractor[reflect.String] = ExtractorValueLen
}

type beEmptyMatcher struct{}

func (m *beEmptyMatcher) ExpectMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	for i, actual := range emptyValueExtractor.ExtractValues(context.Background(), t, actuals...) {
		length := actual.(int)
		if length != 0 {
			t.Fatalf("Expected actual %d with value '%+v' to be empty, but it is not", i, actual)
			panic("unreachable")
		}
	}
	return actuals
}

func (m *beEmptyMatcher) ExpectNoMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	for i, actual := range emptyValueExtractor.ExtractValues(context.Background(), t, actuals...) {
		length := actual.(int)
		if length == 0 {
			t.Fatalf("Expected actual %d with value '%+v' to not be empty, but it is", i, actual)
			panic("unreachable")
		}
	}
	return actuals
}

func BeEmpty() Matcher {
	return &beEmptyMatcher{}
}
