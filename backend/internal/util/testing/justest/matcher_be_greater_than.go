package justest

import (
	"reflect"
)

//go:noinline
func BeGreaterThan(min any) Matcher {
	if min == nil {
		panic("expected a non-nil minimum value")
	}

	return func(t TT, actuals ...any) []any {
		GetHelper(t).Helper()
		var newActuals []any
		for i, actual := range actuals {
			v := numericValueExtractor.MustExtractValue(t, actual)
			actualValue := reflect.ValueOf(v)

			minimumValue := reflect.ValueOf(min)
			if actualValue.Kind() != minimumValue.Kind() {
				t.Fatalf("Expected actual value to be of type '%T', but it is of type '%T'", min, v)
			}

			resultValues := getNumericCompareFuncFor(t, v).Call([]reflect.Value{actualValue, minimumValue})
			if resultValues[0].Int() <= 0 {
				t.Fatalf("Expected actual value %v to be greater than %v", v, min)
			}
			newActuals = append(newActuals, actuals[i])
		}
		return newActuals
	}
}
