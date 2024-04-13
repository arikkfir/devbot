package justest

import "reflect"

//go:noinline
func Fail() Matcher {
	return func(t TT, actuals ...any) []any {
		GetHelper(t).Helper()

		l := len(actuals)
		if l == 0 {
			t.Fatalf("No error occurred")
			panic("unreachable")
		}

		last := actuals[l-1]
		if last == nil {
			t.Fatalf("No error occurred")
			panic("unreachable")
		}

		lastRT := reflect.TypeOf(last)
		if lastRT.AssignableTo(reflect.TypeOf((*error)(nil)).Elem()) {
			// ok, no-op
			return actuals[:l-1]
		}

		t.Fatalf("No error occurred")
		panic("unreachable")
	}
}
