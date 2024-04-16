package justest

import (
	"fmt"
	"time"
)

type Results interface {
	Result() []any
	OrFail()
}

type Evaluator interface {
	Results
	Because(string, ...any) Results
}

type evaluator struct {
	t       TT
	actuals []any
	matcher Matcher
}

func (e *evaluator) Because(format string, args ...any) Results {
	GetHelper(e.t).Helper()
	e.t = &explainedTT{parent: e.t, format: format, args: args}
	return e
}

func (e *evaluator) OrFail() {
	GetHelper(e.t).Helper()
	e.Result()
}

func (e *evaluator) Result() []any {
	GetHelper(e.t).Helper()
	return e.matcher(e.t, e.actuals...)
}

type explainedTT struct {
	parent TT
	format string
	args   []any
}

func (t *explainedTT) Cleanup(f func()) {
	GetHelper(t).Helper()
	t.parent.Cleanup(f)
}

//go:noinline
func (t *explainedTT) Fatalf(format string, args ...any) {
	GetHelper(t).Helper()
	t.parent.Fatalf(format+" (%s)", append(args, fmt.Sprintf(t.format, t.args...))...)
}

//go:noinline
func (t *explainedTT) Log(args ...any) {
	GetHelper(t).Helper()
	t.parent.Log(args...)
}

//go:noinline
func (t *explainedTT) Logf(format string, args ...any) {
	GetHelper(t).Helper()
	t.parent.Logf(format, args...)
}

//go:noinline
func (t *explainedTT) GetHelper() Helper {
	if hp, ok := t.parent.(HelperProvider); ok {
		return hp.GetHelper()
	} else if h, ok := t.parent.(Helper); ok {
		return h
	} else {
		panic(fmt.Sprintf("could not obtain a HelperProvider instance from: %+v", t.parent))
	}
}

//go:noinline
func (t *explainedTT) Deadline() (deadline time.Time, ok bool) {
	GetHelper(t).Helper()
	return t.parent.Deadline()
}

//go:noinline
func (t *explainedTT) Done() <-chan struct{} {
	GetHelper(t).Helper()
	return t.parent.Done()
}

//go:noinline
func (t *explainedTT) Err() error {
	GetHelper(t).Helper()
	return t.parent.Err()
}

//go:noinline
func (t *explainedTT) Value(key any) interface{} {
	GetHelper(t).Helper()
	return getValueForT(t, key)
}