package k8s

import (
	"fmt"
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func HasCondition(conditions []metav1.Condition, conditionType string) bool {
	for _, c := range conditions {
		if c.Type == conditionType {
			return true
		}
	}
	return false
}

func GetCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	for _, c := range conditions {
		if c.Type == conditionType {
			return &c
		}
	}
	return nil
}

func GetConditionReason(conditions []metav1.Condition, conditionType string) string {
	if c := GetCondition(conditions, conditionType); c != nil {
		return c.Reason
	}
	return ""
}

func GetConditionStatus(conditions []metav1.Condition, conditionType string) *metav1.ConditionStatus {
	if c := GetCondition(conditions, conditionType); c != nil {
		status := c.Status
		return &status
	}
	return nil
}

func GetConditionMessage(conditions []metav1.Condition, conditionType string) string {
	if c := GetCondition(conditions, conditionType); c != nil {
		return c.Message
	}
	return ""
}

func IsConditionStatusOneOf(conditions []metav1.Condition, conditionType string, statuses ...metav1.ConditionStatus) bool {
	for _, c := range conditions {
		if c.Type == conditionType {
			for _, status := range statuses {
				if c.Status == status {
					return true
				}
			}
			return false
		}
	}
	return false
}

func RemoveConditionIfReasonIsOneOf(conditions *[]metav1.Condition, conditionType string, reasons ...string) bool {
	changed := false
	var newConditions []metav1.Condition
	if *conditions != nil {
		for _, c := range *conditions {
			if c.Type != conditionType || !slices.Contains(reasons, c.Reason) {
				newConditions = append(newConditions, c)
			} else {
				changed = true
			}
		}
	}
	if changed {
		*conditions = newConditions
	}
	return changed
}

func SetCondition(conditions *[]metav1.Condition, conditionType string, status metav1.ConditionStatus, reason, message string, args ...interface{}) bool {
	if *conditions != nil {
		for i, c := range *conditions {
			if c.Type == conditionType {
				msg := fmt.Sprintf(message, args...)
				if c.Status != status || c.Reason != reason || c.Message != msg {
					(*conditions)[i].Status = status
					(*conditions)[i].Reason = reason
					(*conditions)[i].Message = msg
					return true
				} else {
					return false
				}
			}
		}
	}
	*conditions = append(*conditions, metav1.Condition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func SetConditionsGenerationAndTransitionTime(conditions []metav1.Condition, generation int64) {
	for i, c := range conditions {
		if c.ObservedGeneration == 0 {
			conditions[i].ObservedGeneration = generation
		}
		if c.LastTransitionTime.IsZero() {
			conditions[i].LastTransitionTime = metav1.Now()
		}
	}
}

func ClearStaleConditions(conditions *[]metav1.Condition, currentGeneration int64) {
	if *conditions != nil {
		var newConditions []metav1.Condition
		for _, c := range *conditions {
			if c.ObservedGeneration >= currentGeneration {
				newConditions = append(newConditions, c)
			}
		}
		*conditions = newConditions
	}
}
