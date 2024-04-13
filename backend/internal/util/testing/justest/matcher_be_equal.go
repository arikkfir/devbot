package justest

import (
	"github.com/google/go-cmp/cmp"
	"strings"
)

var (
	equalValueExtractor = NewValueExtractor(ExtractSameValue)
)

type (
	Comparator func(t TT, expected, actual any) any
)

//go:noinline
func BeEqualTo(expected any) Matcher {
	return CompareTo(expected).Using(GoCmp())
}

type ComparatorMatcherBuilder interface {
	Using(Comparator) Matcher
}

type comparatorMatcherBuilder struct {
	expected any
}

//go:noinline
func (c *comparatorMatcherBuilder) Using(comparator Comparator) Matcher {
	return func(t TT, actuals ...any) []any {
		GetHelper(t).Helper()
		var newActuals []any
		for _, actual := range actuals {
			newActuals = append(newActuals, comparator(t, c.expected, equalValueExtractor.MustExtractValue(t, actual)))
		}
		return newActuals
	}
}

//go:noinline
func CompareTo(expected any) ComparatorMatcherBuilder {
	return &comparatorMatcherBuilder{expected: expected}
}

//go:noinline
func GoCmp(opts ...cmp.Option) Comparator {
	return func(t TT, expected, actual any) any {
		GetHelper(t).Helper()
		if cmp.Equal(expected, actual, opts...) {
			return actual
		} else {
			t.Fatalf("Expected & actual value differ:\n%s", strings.TrimSpace(cmp.Diff(expected, actual, opts...)))
			panic("unreachable")
		}
	}
}
