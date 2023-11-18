package v1

import (
	"github.com/secureworks/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

const (
	ConditionTypeCloned = "Cloned"
)

func (o *GitHubRepository) SetStatusConditionClonedIfDifferent(status v1.ConditionStatus, reason, message string) bool {
	object := reflect.ValueOf(o).Elem()

	statusValue := object.FieldByName("Status")
	if !statusValue.IsValid() {
		panic(errors.New("object '%T' does not have a 'Status' field", o))
	}

	conditions := statusValue.FieldByName("Conditions")
	if !conditions.IsValid() {
		panic(errors.New("object '%T' does not have a 'Conditions' field", o))
	}

	for i := 0; i < conditions.Len(); i++ {
		ic := conditions.Index(i).Interface().(v1.Condition)
		if ic.Type == ConditionTypeCloned {
			if ic.Status != status || ic.Reason != reason || ic.Message != message {
				ic.Status = status
				ic.Reason = reason
				ic.Message = message
				ic.LastTransitionTime = v1.Now()
				ic.ObservedGeneration = o.GetGeneration()
				conditions.Index(i).Addr().Elem().Set(reflect.ValueOf(ic))
				return true
			}
			return false
		}
	}
	conditions.Set(reflect.Append(conditions, reflect.ValueOf(v1.Condition{
		Type:               ConditionTypeCloned,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: v1.Now(),
		ObservedGeneration: o.GetGeneration(),
	})))
	return true
}

func (o *GitHubRepository) SetStatusConditionCloned(status v1.ConditionStatus, reason, message string) {
	object := reflect.ValueOf(o).Elem()

	statusValue := object.FieldByName("Status")
	if !statusValue.IsValid() {
		panic(errors.New("object '%T' does not have a 'Status' field", o))
	}

	conditions := statusValue.FieldByName("Conditions")
	if !conditions.IsValid() {
		panic(errors.New("object '%T' does not have a 'Conditions' field", o))
	}

	for i := 0; i < conditions.Len(); i++ {
		ic := conditions.Index(i).Interface().(v1.Condition)
		if ic.Type == ConditionTypeCloned {
			ic.Status = status
			ic.Reason = reason
			ic.Message = message
			ic.LastTransitionTime = v1.Now()
			ic.ObservedGeneration = o.GetGeneration()
			conditions.Index(i).Addr().Elem().Set(reflect.ValueOf(ic))
			return
		}
	}
	conditions.Set(reflect.Append(conditions, reflect.ValueOf(v1.Condition{
		Type:               ConditionTypeCloned,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: v1.Now(),
		ObservedGeneration: o.GetGeneration(),
	})))
}

func (o *GitHubRepository) RemoveStatusConditionCloned() {
	object := reflect.ValueOf(o).Elem()

	statusValue := object.FieldByName("Status")
	if !statusValue.IsValid() {
		panic(errors.New("object '%T' does not have a 'Status' field", o))
	}

	conditions := statusValue.FieldByName("Conditions")
	if !conditions.IsValid() {
		panic(errors.New("object '%T' does not have a 'Conditions' field", o))
	}

	var newConditions []v1.Condition
	for i := 0; i < conditions.Len(); i++ {
		ic := conditions.Index(i).Interface().(v1.Condition)
		if ic.Type != ConditionTypeCloned {
			newConditions = append(newConditions, ic)
		}
	}
	conditions.Set(reflect.ValueOf(newConditions))
}

func (o *GitHubRepository) GetStatusConditionCloned() *v1.Condition {
	object := reflect.ValueOf(o).Elem()

	statusValue := object.FieldByName("Status")
	if !statusValue.IsValid() {
		panic(errors.New("object '%T' does not have a 'Status' field", o))
	}

	conditions := statusValue.FieldByName("Conditions")
	if !conditions.IsValid() {
		panic(errors.New("object '%T' does not have a 'Conditions' field", o))
	}

	for i := 0; i < conditions.Len(); i++ {
		ic := conditions.Index(i).Interface().(v1.Condition)
		if ic.Type == ConditionTypeCloned {
			return &ic
		}
	}
	return nil
}
