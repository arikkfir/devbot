package justest

import (
	"context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
)

var (
	equalConditionValueExtractor = NewValueExtractor(ExtractSameValue)
)

type equalConditionMatcher struct {
	expected *v1.Condition
}

func (m *equalConditionMatcher) ExpectMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	resolvedActuals := equalConditionValueExtractor.ExtractValues(context.Background(), t, actuals...)
	for i, actual := range resolvedActuals {
		actualCondition, ok := actual.(*v1.Condition)
		if !ok {
			t.Fatalf("Expected actual %d value to be a *v1.Condition, but it is not: %+v", i, actual)
		} else if actualCondition == nil && m.expected == nil {
			continue
		} else if actualCondition == nil && m.expected != nil {
			t.Fatalf("Expected condition %d to not be nil, but it is", i)
		} else if actualCondition != nil && m.expected == nil {
			t.Fatalf("Expected condition %d to be nil, but it's not: %+v", i, actualCondition)
		} else if actualCondition.Status != m.expected.Status {
			t.Fatalf("Expected actual %d status to be %s, but it is %s", i, m.expected.Status, actualCondition.Status)
		} else if actualCondition.Reason != m.expected.Reason {
			t.Fatalf("Expected actual %d reason to be %s, but it is %s", i, m.expected.Reason, actualCondition.Reason)
		} else if !regexp.MustCompile(m.expected.Message).Match([]byte(actualCondition.Message)) {
			t.Fatalf("Expected actual %d message to match %s, but it does not: %s", i, m.expected.Message, actualCondition.Message)
		}
	}
	return actuals
}

func (m *equalConditionMatcher) ExpectNoMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()
	resolvedActuals := equalConditionValueExtractor.ExtractValues(context.Background(), t, actuals...)
	for i, actual := range resolvedActuals {
		actualCondition, ok := actual.(*v1.Condition)
		if !ok {
			t.Fatalf("Expected actual %d value to be a *v1.Condition, but it is not: %+v", i, actual)
		} else if actualCondition == nil && m.expected == nil {
			t.Fatalf("Expected actual %d message not to match: %+v", i, m.expected)
		} else if actualCondition == nil && m.expected != nil {
			continue
		} else if actualCondition != nil && m.expected == nil {
			continue
		} else if actualCondition.Status != m.expected.Status {
			continue
		} else if actualCondition.Reason != m.expected.Reason {
			continue
		} else if regexp.MustCompile(m.expected.Message).Match([]byte(actualCondition.Message)) {
			continue
		} else {
			t.Fatalf("Expected actual %d message not to match: %+v", i, m.expected)
		}
	}
	return actuals
}

func EqualCondition(expected *v1.Condition) Matcher {
	return &equalConditionMatcher{expected: expected}
}
