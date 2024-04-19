package justest

import (
	"github.com/secureworks/errors"
	"strings"
)

type Results interface {
	Result() []any
	OrFail()
}

type Evaluator interface {
	Results
	Because(string, ...any) Results
}

//go:noinline
func newEvaluator(t TT, matcher Matcher, actuals ...any) Evaluator {
	GetHelper(t).Helper()
	e := &evaluator{t: t, actuals: actuals, matcher: matcher}

	loc := unevaluatedLocation{function: "unknown", file: "unknown", line: 0}
	for _, frame := range errors.CallStackAt(0) {
		function, file, line := frame.Location()

		startsWithAnIgnoredPrefix := false
		for _, prefix := range ignoredStackTracePrefixes {
			if strings.HasPrefix(function, prefix) {
				startsWithAnIgnoredPrefix = true
				break
			}
		}

		if !startsWithAnIgnoredPrefix {
			loc = unevaluatedLocation{function: function, file: file, line: line}
		}
	}

	t.Cleanup(func() {
		GetHelper(t).Helper()
		if !e.evaluated {
			unevaluated.Add(t, loc.function, loc.file, loc.line)
		}
	})

	return e
}

type evaluator struct {
	t         TT
	actuals   []any
	matcher   Matcher
	evaluated bool
}

//go:noinline
func (e *evaluator) Because(format string, args ...any) Results {
	GetHelper(e.t).Helper()
	et := &explainedTT{parent: e.t, format: format, args: args}
	e.t.Cleanup(et.PerformCleanups)
	e.t = et
	return e
}

//go:noinline
func (e *evaluator) OrFail() {
	GetHelper(e.t).Helper()
	e.Result()
}

//go:noinline
func (e *evaluator) Result() []any {
	GetHelper(e.t).Helper()
	e.evaluated = true
	result := e.matcher(e.t, e.actuals...)
	return result
}
