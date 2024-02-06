//go:build !ignore_autogenerated

// Code generated by devbot script. DO NOT EDIT.

package v1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"slices"
)

func (s *ObjectWithoutCommonConditionsStatus) SetInvalidDueToC1(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != C1 || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = C1
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  C1,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *ObjectWithoutCommonConditionsStatus) SetMaybeInvalidDueToC1(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != C1 || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = C1
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  C1,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *ObjectWithoutCommonConditionsStatus) SetInvalidDueToC2(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != C2 || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = C2
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  C2,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *ObjectWithoutCommonConditionsStatus) SetMaybeInvalidDueToC2(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != C2 || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = C2
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  C2,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *ObjectWithoutCommonConditionsStatus) SetValidIfInvalidDueToAnyOf(reasons ...string) bool {
	changed := false
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || !slices.Contains(reasons, c.Reason) {
			newConditions = append(newConditions, c)
		} else {
			changed = true
		}
	}
	if changed {
		s.Conditions = newConditions
	}
	return changed
}

func (s *ObjectWithoutCommonConditionsStatus) SetValid() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *ObjectWithoutCommonConditionsStatus) IsValid() bool {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Status != v1.ConditionTrue
		}
	}
	return true
}

func (s *ObjectWithoutCommonConditionsStatus) IsInvalid() bool {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *ObjectWithoutCommonConditionsStatus) GetInvalidCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *ObjectWithoutCommonConditionsStatus) GetInvalidReason() string {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Reason
		}
	}
	return ""
}

func (s *ObjectWithoutCommonConditionsStatus) GetInvalidStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *ObjectWithoutCommonConditionsStatus) GetInvalidMessage() string {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Message
		}
	}
	return ""
}

func (s *ObjectWithoutCommonConditionsStatus) GetConditions() []v1.Condition {
	return s.Conditions
}

func (s *ObjectWithoutCommonConditionsStatus) SetConditions(conditions []v1.Condition) {
	s.Conditions = conditions
}

func (s *ObjectWithoutCommonConditionsStatus) ClearStaleConditions(currentGeneration int64) {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.ObservedGeneration >= currentGeneration {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}