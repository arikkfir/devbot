package justest

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"regexp"
)

var (
	sayValueExtractor ValueExtractor
)

func init() {
	sayValueExtractor = NewValueExtractor(ExtractorUnsupported)
	sayValueExtractor[reflect.Chan] = NewChannelExtractor(sayValueExtractor, true)
	sayValueExtractor[reflect.Func] = NewFuncExtractor(sayValueExtractor, true)
	sayValueExtractor[reflect.Pointer] = func(ctx context.Context, t JustT, v any) (any, bool) {
		GetHelper(t).Helper()
		if bufferPointer, ok := v.(*bytes.Buffer); ok {
			return bufferPointer.String(), true
		} else if stringPointer, ok := v.(*string); ok {
			return *stringPointer, true
		} else if ba, ok := v.(*[]byte); ok {
			return string(*ba), true
		} else {
			t.Fatalf("Unsupported type '%T' for Say matcher: %+v", v, v)
			panic("unreachable")
		}
	}
	sayValueExtractor[reflect.Slice] = func(ctx context.Context, t JustT, v any) (any, bool) {
		GetHelper(t).Helper()
		if b, ok := v.([]byte); ok {
			return string(b), true
		} else {
			t.Fatalf("Unsupported type '%T' for Say matcher: %+v", v, v)
			panic("unreachable")
		}
	}
	sayValueExtractor[reflect.String] = ExtractSameValue
}

type sayMatcher struct {
	re *regexp.Regexp
}

func (m *sayMatcher) ExpectMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	for i, actual := range sayValueExtractor.ExtractValues(context.Background(), t, actuals...) {
		if !m.re.Match([]byte(actual.(string))) {
			t.Fatalf("Expected actual %d to match '%s', but it does not: %s", i, m.re, actual)
		}
	}
	return actuals
}

func (m *sayMatcher) ExpectNoMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	for i, actual := range sayValueExtractor.ExtractValues(context.Background(), t, actuals...) {
		if m.re.Match([]byte(actual.(string))) {
			t.Fatalf("Expected actual %d not to match '%s', but it does: %s", i, m.re, actual)
		}
	}
	return actuals
}

func Say[T string | *regexp.Regexp](expectation T) Matcher {
	switch t := any(expectation).(type) {
	case string:
		return &sayMatcher{re: regexp.MustCompile(t)}
	case *regexp.Regexp:
		return &sayMatcher{re: t}
	default:
		panic(fmt.Sprintf("unsupported type for Say matcher: %T", expectation))
	}
}
