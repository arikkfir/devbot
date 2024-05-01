package justest

import (
	"github.com/google/go-cmp/cmp"
)

type Comparator func(t T, expected any, actual any)

type EqualToMatcher interface {
	Matcher
	Using(comparator Comparator) EqualToMatcher
}

type equalTo struct {
	expected   []any
	comparator Comparator
}

//go:noinline
func (m *equalTo) Assert(t T, actuals ...any) {
	GetHelper(t).Helper()
	if len(m.expected) != len(actuals) {
		t.Fatalf("Unexpected difference: received %d actual values and %d expected values", len(actuals), len(m.expected))
	} else {
		for i := 0; i < len(m.expected); i++ {
			expected := m.expected[i]
			actual := actuals[i]
			m.comparator(t, expected, actual)
		}
	}
}

func (m *equalTo) Using(comparator Comparator) EqualToMatcher {
	m.comparator = comparator
	return m
}

//go:noinline
func EqualTo(expected ...any) EqualToMatcher {
	return &equalTo{
		expected: expected,
		comparator: func(t T, expected, actual any) {
			if !cmp.Equal(expected, actual) {
				t.Fatalf("Unexpected difference:\n%s", cmp.Diff(expected, actual))
			}
		},
	}
}