package v1

import (
	"fmt"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *ApplicationStatus) SetStaleDueToInternalError(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = InternalError
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  InternalError,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetMaybeStaleDueToInternalError(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = InternalError
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  InternalError,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetCurrentIfStaleDueToInternalError() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != InternalError {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *ApplicationStatus) SetStaleDueToInvalid(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = Invalid
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  Invalid,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetMaybeStaleDueToInvalid(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = Invalid
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  Invalid,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetCurrentIfStaleDueToInvalid() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != Invalid {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *ApplicationStatus) SetStaleDueToRepositoryNotAccessible(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = RepositoryNotAccessible
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryNotAccessible,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetMaybeStaleDueToRepositoryNotAccessible(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = RepositoryNotAccessible
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryNotAccessible,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetCurrentIfStaleDueToRepositoryNotAccessible() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != RepositoryNotAccessible {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *ApplicationStatus) SetStaleDueToRepositoryNotFound(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = RepositoryNotFound
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryNotFound,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetMaybeStaleDueToRepositoryNotFound(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = RepositoryNotFound
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryNotFound,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetCurrentIfStaleDueToRepositoryNotFound() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != RepositoryNotFound {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *ApplicationStatus) SetCurrent() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *ApplicationStatus) IsCurrent() bool {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Status == v1.ConditionTrue
		}
	}
	return true
}

func (s *ApplicationStatus) IsStale() bool {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *ApplicationStatus) GetStaleCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *ApplicationStatus) GetStaleReason() string {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Reason
		}
	}
	return ""
}

func (s *ApplicationStatus) GetStaleStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *ApplicationStatus) GetStaleMessage() string {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Message
		}
	}
	return ""
}

func (s *ApplicationStatus) SetInvalidDueToAddFinalizerFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = AddFinalizerFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  AddFinalizerFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetMaybeInvalidDueToAddFinalizerFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = AddFinalizerFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  AddFinalizerFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetValidIfInvalidDueToAddFinalizerFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != AddFinalizerFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *ApplicationStatus) SetInvalidDueToControllerMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = ControllerMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  ControllerMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetMaybeInvalidDueToControllerMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = ControllerMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  ControllerMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetValidIfInvalidDueToControllerMissing() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != ControllerMissing {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *ApplicationStatus) SetInvalidDueToFailedGettingOwnedObjects(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = FailedGettingOwnedObjects
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  FailedGettingOwnedObjects,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetMaybeInvalidDueToFailedGettingOwnedObjects(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = FailedGettingOwnedObjects
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  FailedGettingOwnedObjects,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetValidIfInvalidDueToFailedGettingOwnedObjects() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != FailedGettingOwnedObjects {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *ApplicationStatus) SetInvalidDueToFinalizationFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = FinalizationFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  FinalizationFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetMaybeInvalidDueToFinalizationFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = FinalizationFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  FinalizationFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetValidIfInvalidDueToFinalizationFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != FinalizationFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *ApplicationStatus) SetInvalidDueToInternalError(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = InternalError
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  InternalError,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetMaybeInvalidDueToInternalError(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = InternalError
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  InternalError,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetValidIfInvalidDueToInternalError() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != InternalError {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *ApplicationStatus) SetInvalidDueToRepositoryNotSupported(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = RepositoryNotSupported
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryNotSupported,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetMaybeInvalidDueToRepositoryNotSupported(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = RepositoryNotSupported
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryNotSupported,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *ApplicationStatus) SetValidIfInvalidDueToRepositoryNotSupported() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != RepositoryNotSupported {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *ApplicationStatus) SetValid() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *ApplicationStatus) IsValid() bool {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Status == v1.ConditionTrue
		}
	}
	return true
}

func (s *ApplicationStatus) IsInvalid() bool {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *ApplicationStatus) GetInvalidCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *ApplicationStatus) GetInvalidReason() string {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Reason
		}
	}
	return ""
}

func (s *ApplicationStatus) GetInvalidStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *ApplicationStatus) GetInvalidMessage() string {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Message
		}
	}
	return ""
}

func (s *ApplicationStatus) GetConditions() []v1.Condition {
	return s.Conditions
}

func (s *ApplicationStatus) SetConditions(conditions []v1.Condition) {
	s.Conditions = conditions
}

func (s *ApplicationStatus) ClearStaleConditions(currentGeneration int64) {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.ObservedGeneration >= currentGeneration {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (o *Application) GetStatus() k8s.CommonStatus {
	return &o.Status
}
