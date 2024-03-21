package justest

import (
	"fmt"
	"reflect"
)

var (
	justTTypePkgPath string
	justTTypeName    string
)

func init() {
	justTType := reflect.TypeOf((*JustT)(nil)).Elem()
	justTTypePkgPath = justTType.PkgPath()
	justTTypeName = justTType.Name()
}

type JustT interface {
	Cleanup(func())
	Fatalf(format string, args ...any)
	Log(args ...any)
	Logf(format string, args ...any)
}

type Helper interface {
	Helper()
}

type HelperProvider interface {
	GetHelper() Helper
}

func GetHelper(t JustT) Helper {
	if hp, ok := t.(HelperProvider); ok {
		return hp.GetHelper()
	} else if h, ok := t.(Helper); ok {
		return h
	} else {
		panic(fmt.Sprintf("could not obtain a HelperProvider instance from the given JustT instance: %+v", t))
	}
}

func For(t JustT) Tester {
	GetHelper(t).Helper()
	return &tester{t: t}
}

type Tester interface {
	Expect(actual ...any) Asserter
}

type tester struct {
	t JustT
}

func (t *tester) Expect(actual ...interface{}) Asserter {
	GetHelper(t.t).Helper()
	return &asserter{t: t, actual: actual}
}

type Asserter interface {
	Will(Matcher) ContinueAsserter
	WillNot(Matcher) ContinueAsserter
	Results() []any
}

type ContinueAsserter interface {
	AndWill(Matcher) ContinueAsserter
	AndWillNot(Matcher) ContinueAsserter
	Results() []any
}

type asserter struct {
	t      *tester
	actual []any
}

func (a *asserter) Will(m Matcher) ContinueAsserter {
	GetHelper(a.t.t).Helper()
	newActuals := m.ExpectMatch(a.t.t, a.actual...)
	return &asserter{t: a.t, actual: newActuals}
}

func (a *asserter) WillNot(m Matcher) ContinueAsserter {
	GetHelper(a.t.t).Helper()
	newActuals := m.ExpectNoMatch(a.t.t, a.actual...)
	return &asserter{t: a.t, actual: newActuals}
}

func (a *asserter) AndWill(m Matcher) ContinueAsserter {
	GetHelper(a.t.t).Helper()
	return a.Will(m)
}

func (a *asserter) AndWillNot(m Matcher) ContinueAsserter {
	GetHelper(a.t.t).Helper()
	return a.WillNot(m)
}

func (a *asserter) Results() []any {
	GetHelper(a.t.t).Helper()
	return a.actual
}

type Matcher interface {
	ExpectMatch(t JustT, actuals ...any) []any
	ExpectNoMatch(t JustT, actuals ...any) []any
}
