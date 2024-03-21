package justest

import (
	"context"
	"reflect"
)

type beLessThanMatcher struct {
	max any
}

func (m *beLessThanMatcher) ExpectMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	for i, actual := range numericValueExtractor.ExtractValues(context.Background(), t, actuals...) {
		actualValue := reflect.ValueOf(actual)

		maximumValue := reflect.ValueOf(m.max)
		if actualValue.Kind() != maximumValue.Kind() {
			t.Fatalf("Expected actual %d with value %v to be of type '%T' (like the maximum value's type), but it is of type '%T'", i, actual, m.max, actual)
		}

		cmpCompareFunctionValue := getNumericCompareFuncFor(t, actual)

		resultValues := cmpCompareFunctionValue.Call([]reflect.Value{actualValue, maximumValue})
		if resultValues[0].Int() >= 0 {
			t.Fatalf("Expected actual %d with value %v to be less than %v", i, actual, m.max)
		}
	}
	return actuals
}

func (m *beLessThanMatcher) ExpectNoMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	for i, actual := range numericValueExtractor.ExtractValues(context.Background(), t, actuals...) {
		actualValue := reflect.ValueOf(actual)

		maximumValue := reflect.ValueOf(m.max)
		if actualValue.Kind() != maximumValue.Kind() {
			t.Fatalf("Expected actual %d with value %v to be of type '%T' (like the maximum value's type), but it is of type '%T'", i, actual, m.max, actual)
		}

		cmpCompareFunctionValue := getNumericCompareFuncFor(t, actual)

		resultValues := cmpCompareFunctionValue.Call([]reflect.Value{actualValue, maximumValue})
		if resultValues[0].Int() < 0 {
			t.Fatalf("Expected actual %d with value %v to not be less than %v", i, actual, m.max)
		}
	}
	return actuals
}

func BeLessThan(max any) Matcher {
	if max == nil {
		panic("expected a non-nil maximum value")
	}
	return &beLessThanMatcher{max: max}
}
