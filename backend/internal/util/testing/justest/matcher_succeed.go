package justest

import "reflect"

//go:noinline
func Succeed() Matcher {
	return func(t TT, actuals ...any) []any {
		GetHelper(t).Helper()

		l := len(actuals)
		if l == 0 {
			return actuals
		}

		last := actuals[l-1]
		if last == nil {
			return actuals[:l-1]
		}

		lastRT := reflect.TypeOf(last)
		if lastRT.AssignableTo(reflect.TypeOf((*error)(nil)).Elem()) {
			t.Fatalf("Error occurred: %+v", last)
		}

		return actuals[:l-1]
	}
}
