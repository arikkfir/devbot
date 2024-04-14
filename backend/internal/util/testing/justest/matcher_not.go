package justest

import (
	"fmt"
	"reflect"
	"runtime"
	"time"
)

type inverseTT struct {
	parent    TT
	fatalMsg  *string
	fatalArgs *[]any
}

//go:noinline
func (t *inverseTT) Cleanup(f func()) {
	GetHelper(t.parent).Helper()
	t.parent.Cleanup(f)
}

//go:noinline
func (t *inverseTT) Fatalf(format string, args ...any) {
	GetHelper(t.parent).Helper()
	f := format
	a := args
	t.fatalMsg = &f
	t.fatalArgs = &a
}

//go:noinline
func (t *inverseTT) Log(args ...any) {
	GetHelper(t.parent).Helper()
	t.parent.Log(args...)
}

//go:noinline
func (t *inverseTT) Logf(format string, args ...any) {
	GetHelper(t.parent).Helper()
	t.parent.Logf(format, args...)
}

//go:noinline
func (t *inverseTT) GetHelper() Helper {
	if hp, ok := t.parent.(HelperProvider); ok {
		return hp.GetHelper()
	} else if h, ok := t.parent.(Helper); ok {
		return h
	} else {
		panic(fmt.Sprintf("could not obtain a HelperProvider instance from: %+v", t.parent))
	}
}

//go:noinline
func (t *inverseTT) Deadline() (deadline time.Time, ok bool) {
	GetHelper(t).Helper()
	return t.parent.Deadline()
}

//go:noinline
func (t *inverseTT) Done() <-chan struct{} { GetHelper(t).Helper(); return t.parent.Done() }

//go:noinline
func (t *inverseTT) Err() error { GetHelper(t).Helper(); return t.parent.Err() }

//go:noinline
func (t *inverseTT) Value(key any) interface{} { GetHelper(t).Helper(); return getValueForT(t, key) }

//go:noinline
func Not(m Matcher) Matcher {
	return func(t TT, actuals ...any) []any {
		GetHelper(t).Helper()

		var newActuals []any

		handleActual := func(t TT, actual any) any {
			tt := &inverseTT{parent: t}

			defer func() {
				if r := recover(); r != nil {
					if tt.fatalMsg != nil {
						// no-op: a call to tt.Fatalf(...) is expected and wanted, do nothing
					} else {
						// an unexpected panic! re-panic
						panic(r)
					}
				}
			}()

			_ = m(tt, actual)
			funcForPC := runtime.FuncForPC(reflect.ValueOf(m).Pointer())
			t.Fatalf("Expected match failure, which did not occur!\nMatcher: %+v\nActual value: %+v", funcForPC.Name(), actual)
			panic("unreachable")
		}
		for _, actual := range actuals {
			actual := actual
			newActuals = append(newActuals, handleActual(t, actual))
		}
		return newActuals
	}
}
