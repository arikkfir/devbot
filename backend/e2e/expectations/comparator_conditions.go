package expectations

import (
	. "github.com/arikkfir/justest"
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
	actual := a.(*metav1.Condition)
	expectation := e.(*ConditionE)
	if expectation != nil {
		With(t).Verify(actual).Will(Not(BeNil())).OrFail()
		With(t).Verify(actual.Type).Will(EqualTo(expectation.Type)).OrFail()
		if expectation.Status != nil {
			With(t).Verify(actual.Status).Will(EqualTo(metav1.ConditionStatus(*expectation.Status))).OrFail()
		}
		if expectation.Reason != nil {
			With(t).Verify(actual.Reason).Will(Say(expectation.Reason)).OrFail()
		}
		if expectation.Message != nil {
			With(t).Verify(actual.Message).Will(Say(expectation.Message)).OrFail()
		}
	} else {
		With(t).Verify(actual).Will(BeNil()).OrFail()
	}
}
