package justest

import (
	"context"
	"reflect"
)

type beBetweenMatcher struct {
	min any
	max any
}

func (m *beBetweenMatcher) ExpectMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	for i, actual := range numericValueExtractor.ExtractValues(context.Background(), t, actuals...) {
		actualValue := reflect.ValueOf(actual)

		minimumValue := reflect.ValueOf(m.min)
		if actualValue.Kind() != minimumValue.Kind() {
			t.Fatalf("Expected actual %d with value %v to be of type '%T' (like the minimum value's type), but it is of type '%T'", i, actual, m.min, actual)
		}

		maximumValue := reflect.ValueOf(m.max)
		if actualValue.Kind() != maximumValue.Kind() {
			t.Fatalf("Expected actual %d with value %v to be of type '%T' (like the maximum value's type), but it is of type '%T'", i, actual, m.max, actual)
		}

		cmpCompareFunctionValue := getNumericCompareFuncFor(t, actual)

		resultActualVsMin := cmpCompareFunctionValue.Call([]reflect.Value{actualValue, minimumValue})
		if result := resultActualVsMin[0].Int(); result < 0 {
			t.Fatalf("Expected actual %d with value %v to be between %v and %v", i, actual, m.min, m.max)
		}

		resultActualVsMax := cmpCompareFunctionValue.Call([]reflect.Value{actualValue, maximumValue})
		if result := resultActualVsMax[0].Int(); result > 0 {
			t.Fatalf("Expected actual %d with value %v to be between %v and %v", i, actual, m.min, m.max)
		}
	}
	return actuals
}

func (m *beBetweenMatcher) ExpectNoMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	for i, actual := range numericValueExtractor.ExtractValues(context.Background(), t, actuals...) {
		actualValue := reflect.ValueOf(actual)

		minimumValue := reflect.ValueOf(m.min)
		if actualValue.Kind() != minimumValue.Kind() {
			t.Fatalf("Expected actual %d with value %v to be of type '%T' (like the minimum value's type), but it is of type '%T'", i, actual, m.min, actual)
		}

		maximumValue := reflect.ValueOf(m.max)
		if actualValue.Kind() != maximumValue.Kind() {
			t.Fatalf("Expected actual %d with value %v to be of type '%T' (like the maximum value's type), but it is of type '%T'", i, actual, m.max, actual)
		}

		cmpCompareFunctionValue := getNumericCompareFuncFor(t, actual)

		minComparisonResult := cmpCompareFunctionValue.Call([]reflect.Value{actualValue, minimumValue})[0].Int()
		if minComparisonResult >= 0 {
			maxComparisonResult := cmpCompareFunctionValue.Call([]reflect.Value{actualValue, maximumValue})[0].Int()
			if maxComparisonResult <= 0 {
				t.Fatalf("Expected actual %d with value %v not to be between %v and %v", i, actual, m.min, m.max)
				panic("unreachable")
			}
		}
	}
	return actuals
}

func BeBetween(min, max any) Matcher {
	if min == nil {
		panic("expected a non-nil minimum value")
	} else if max == nil {
		panic("expected a non-nil maximum value")
	}
	return &beBetweenMatcher{min: min, max: max}
}
