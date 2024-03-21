package justest

import (
	"context"
	"reflect"
)

type beGreaterThanMatcher struct {
	min any
}

func (m *beGreaterThanMatcher) ExpectMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	for i, actual := range numericValueExtractor.ExtractValues(context.Background(), t, actuals...) {
		actualValue := reflect.ValueOf(actual)

		minimumValue := reflect.ValueOf(m.min)
		if actualValue.Kind() != minimumValue.Kind() {
			t.Fatalf("Expected actual %d with value %v to be of type '%T' (like the minimum value's type), but it is of type '%T'", i, actual, m.min, actual)
		}

		cmpCompareFunctionValue := getNumericCompareFuncFor(t, actual)

		resultValues := cmpCompareFunctionValue.Call([]reflect.Value{actualValue, minimumValue})
		if resultValues[0].Int() <= 0 {
			t.Fatalf("Expected actual %d with value %v to be greater than %v", i, actual, m.min)
		}
	}
	return actuals
}

func (m *beGreaterThanMatcher) ExpectNoMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	for i, actual := range numericValueExtractor.ExtractValues(context.Background(), t, actuals...) {
		actualValue := reflect.ValueOf(actual)

		minimumValue := reflect.ValueOf(m.min)
		if actualValue.Kind() != minimumValue.Kind() {
			t.Fatalf("Expected actual %d with value %v to be of type '%T' (like the minimum value's type), but it is of type '%T'", i, actual, m.min, actual)
		}

		cmpCompareFunctionValue := getNumericCompareFuncFor(t, actual)

		resultValues := cmpCompareFunctionValue.Call([]reflect.Value{actualValue, minimumValue})
		if resultValues[0].Int() > 0 {
			t.Fatalf("Expected actual %d with value %v to not be greater than %v", i, actual, m.min)
		}
	}
	return actuals
}

func BeGreaterThan(min any) Matcher {
	if min == nil {
		panic("expected a non-nil minimum value")
	}
	return &beGreaterThanMatcher{min: min}
}
