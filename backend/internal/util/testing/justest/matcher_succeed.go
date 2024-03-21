package justest

import (
	"context"
)

var (
	succeedValueExtractor = NewValueExtractor(ExtractSameValue)
)

type succeedMatcher struct{}

func (m *succeedMatcher) ExpectMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	resolvedActuals := succeedValueExtractor.ExtractValues(context.Background(), t, actuals...)
	if len(resolvedActuals) > 0 {
		last := resolvedActuals[len(resolvedActuals)-1]
		if last != nil {
			if err, ok := last.(error); ok {
				t.Fatalf("Expected last actual value to be nil, missing or not an error - but it is a non-nil error: %+v", err)
			}
		}
	}
	return actuals
}

func (m *succeedMatcher) ExpectNoMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	resolvedActuals := succeedValueExtractor.ExtractValues(context.Background(), t, actuals...)
	if len(resolvedActuals) > 0 {
		last := resolvedActuals[len(resolvedActuals)-1]
		if last != nil {
			if _, ok := last.(error); ok {
				return actuals
			}
		}
	}
	t.Fatalf("Expected last actual value to be an error - but it is missing, nil, or not an error")
	panic("unreachable")
}

func Succeed() Matcher {
	return &succeedMatcher{}
}
