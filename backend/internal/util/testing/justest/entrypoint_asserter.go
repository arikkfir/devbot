package justest

type Asserter interface {
	Will(Matcher) Evaluator
	WillNot(Matcher) Evaluator
}

type asserter struct {
	t       TT
	actuals []any
}

//go:noinline
func (a *asserter) Will(matcher Matcher) Evaluator {
	GetHelper(a.t).Helper()
	return newEvaluator(a.t, matcher, a.actuals...)
}

//go:noinline
func (a *asserter) WillNot(matcher Matcher) Evaluator {
	GetHelper(a.t).Helper()
	return newEvaluator(a.t, Not(matcher), a.actuals...)
}
