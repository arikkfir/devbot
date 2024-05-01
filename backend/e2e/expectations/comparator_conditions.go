package expectations

import (
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ConditionsComparator(t T, e, a any) {
	expectedConditions := e.(map[string]*ConditionE)
	actualConditions := append([]metav1.Condition{}, a.([]metav1.Condition)...) // cloning so we can chip away at it

	for expectedConditionType, expectedConditionProperties := range expectedConditions {
		found := false
		for i, actualCondition := range actualConditions {
			if actualCondition.Type == expectedConditionType {
				found = true
				With(t).Verify(&actualCondition).Will(EqualTo(expectedConditionProperties).Using(ConditionComparator)).OrFail()
				actualConditions = append(actualConditions[:i], actualConditions[i+1:]...)
				break
			}
		}
		if expectedConditionProperties != nil {
			With(t).Verify(found).Will(EqualTo(true)).OrFail()
		} else {
			With(t).Verify(found).Will(EqualTo(false)).OrFail()
		}
	}
	With(t).Verify(len(actualConditions)).Will(EqualTo(0)).OrFail()
}

func ConditionComparator(t T, e, a any) {
	expected := e.(*ConditionE)
	actual := a.(*metav1.Condition)

	if actual == nil && expected == nil {
		return
	} else if actual == nil && expected != nil {
		t.Fatalf("Expected condition '%s' to exist, but it does not", expected.Type)
	} else if actual != nil && expected == nil {
		t.Fatalf("Expected condition '%s' not to exist, but it does", actual.Type)
	} else if actual.Type != expected.Type {
		t.Fatalf("Expected condition '%s', but got '%s'", expected.Type, actual.Type)
	}

	if expected.Status != nil {
		With(t).Verify(string(actual.Status)).Will(EqualTo(*expected.Status)).OrFail()
	}
	if expected.Reason != nil {
		With(t).Verify(actual.Reason).Will(Say(expected.Reason)).OrFail()
	}
	if expected.Message != nil {
		With(t).Verify(actual.Message).Will(Say(expected.Message)).OrFail()
	}
}
