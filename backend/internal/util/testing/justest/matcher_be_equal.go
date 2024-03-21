package justest

import (
	"context"
	"github.com/google/go-cmp/cmp"
)

var (
	equalValueExtractor = NewValueExtractor(ExtractSameValue)
)

type beEqualMatcherTo struct {
	expected []any
	cmpOpts  []cmp.Option
}

func (m *beEqualMatcherTo) ExpectMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	resolvedActuals := equalValueExtractor.ExtractValues(context.Background(), t, actuals...)
	if cmp.Equal(m.expected, resolvedActuals, m.cmpOpts...) {
		return actuals
	} else {
		t.Fatalf("Expected actual values to equal expected values, but they do not:\n%s", cmp.Diff(m.expected, resolvedActuals))
		panic("unreachable")
	}
}

func (m *beEqualMatcherTo) ExpectNoMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	resolvedActuals := equalValueExtractor.ExtractValues(context.Background(), t, actuals...)
	if cmp.Equal(m.expected, resolvedActuals, m.cmpOpts...) {
		t.Fatalf("Expected actual values not to equal expected values, but they do:\n--[ Expected: ]--\n%+v\n--[ Actual: ]--\n%+v", m.expected, resolvedActuals)
		panic("unreachable")
	} else {
		return actuals
	}
}

func BeEqualTo(expected ...any) Matcher {
	var cmpOpts []cmp.Option
	var expectedMinusCmpOpts []any
	for _, e := range expected {
		e := e
		if _, ok := e.(cmp.Option); ok {
			cmpOpts = append(cmpOpts, e.(cmp.Option))
		} else {
			expectedMinusCmpOpts = append(expectedMinusCmpOpts, e)
		}
	}
	return &beEqualMatcherTo{expected: expectedMinusCmpOpts, cmpOpts: cmpOpts}
}
