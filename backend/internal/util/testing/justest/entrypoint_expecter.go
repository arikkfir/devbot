package justest

import "context"

//go:noinline
func For(t T) Expecter {
	GetHelper(t).Helper()
	tt := getValueForT(t, "___TT")
	if tt == nil {
		tt = NewTT(t)
		setValueForT(t, "___TT", tt)
	}
	return &expecter{t: tt.(TT)}
}

type Expecter interface {
	AddValue(key, value any)
	Value(key any) any
	Context() context.Context
	Expect(actual ...any) Asserter
}

type expecter struct {
	t TT
}

//go:inline
func (e *expecter) Context() context.Context {
	GetHelper(e.t).Helper()
	return e.t.(context.Context)
}

//go:noinline
func (e *expecter) AddValue(key, value any) {
	GetHelper(e.t).Helper()
	setValueForT(e.t, key, value)
}

//go:noinline
func (e *expecter) Value(key any) any {
	GetHelper(e.t).Helper()
	return getValueForT(e.t, key)
}

//go:noinline
func (e *expecter) Expect(actuals ...any) Asserter {
	GetHelper(e.t).Helper()
	return &asserter{t: e.t, actuals: actuals}
}
