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
	return &evaluator{t: a.t, actuals: a.actuals, matcher: matcher}
}

//go:noinline
func (a *asserter) WillNot(matcher Matcher) Evaluator {
	return a.Will(Not(matcher))
}
