package justest

import (
	"context"
	"reflect"
)

var (
	nilValueExtractor = NewValueExtractor(ExtractSameValue)
)

type beNilMatcher struct{}

func (m *beNilMatcher) ExpectMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	for i, actual := range nilValueExtractor.ExtractValues(context.Background(), t, actuals...) {
		switch v := reflect.ValueOf(actual); v.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			if !v.IsNil() {
				t.Fatalf("Expected actual %d to be nil, but it is not: %+v", i, actual)
				panic("unreachable")
			}
		default:
			if actual != nil {
				t.Fatalf("Expected actual %d to be nil, but it is not: %+v", i, actual)
				panic("unreachable")
			}
		}
	}
	return actuals
}

func (m *beNilMatcher) ExpectNoMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	for i, actual := range nilValueExtractor.ExtractValues(context.Background(), t, actuals...) {
		switch v := reflect.ValueOf(actual); v.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			if v.IsNil() {
				t.Fatalf("Expected actual %d to not be nil, but it is", i)
				panic("unreachable")
			}
		default:
			if actual == nil {
				t.Fatalf("Expected actual %d to not be nil, but it is", i)
				panic("unreachable")
			}
		}
	}
	return actuals
}

func BeNil() Matcher {
	return &beNilMatcher{}
}
