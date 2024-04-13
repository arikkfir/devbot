package justest

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

var (
	contextValues = sync.Map{}
)

func getValueForT(t T, key any) any {
	GetHelper(t).Helper()
	values, ok := contextValues.Load(t)
	if !ok {
		switch tt := t.(type) {
		case *tImpl:
			return tt.ctx.Value(key)
		case *inverseTT:
			return tt.parent.Value(key)
		case *eventuallyT:
			return tt.parent.Value(key)
		case *MockT:
			return tt.Parent.Value(key)
		case *testing.T:
			return nil
		default:
			panic(fmt.Sprintf("unrecognized TT type: %T", t))
		}
	}
	return values.(map[any]any)[key]
}

func setValueForT(t T, key, value any) {
	GetHelper(t).Helper()
	values, _ := contextValues.LoadOrStore(t, make(map[any]any))
	if value == nil {
		delete(values.(map[any]any), key)
	} else {
		values.(map[any]any)[key] = value
	}
}

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
	return &asserter{t: e, actuals: actuals}
}

type Asserter interface {
	Will(Matcher) ContinueAsserter
	Result() []any
}

type ContinueAsserter interface {
	AndWill(Matcher) ContinueAsserter
	Result() []any
}

type asserter struct {
	t       *expecter
	actuals []any
	because string
}

//go:noinline
func (a *asserter) Will(matcher Matcher) ContinueAsserter {
	GetHelper(a.t.t).Helper()
	newActuals := matcher(a.t.t, a.actuals...)
	return &asserter{t: a.t, actuals: newActuals}
}

func (a *asserter) Because(format string, args ...any) ContinueAsserter {
	GetHelper(a.t.t).Helper()
	a.because = fmt.Sprintf(format, args...)
	return a
}

//go:noinline
func (a *asserter) AndWill(m Matcher) ContinueAsserter {
	GetHelper(a.t.t).Helper()
	return a.Will(m)
}

//go:noinline
func (a *asserter) Result() []any {
	GetHelper(a.t.t).Helper()
	return a.actuals
}

type Matcher func(TT, ...any) []any
