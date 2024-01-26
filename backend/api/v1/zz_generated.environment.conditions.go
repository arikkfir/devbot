package v1

import (
	"fmt"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *EnvironmentStatus) SetStaleDueToInternalError(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetMaybeStaleDueToInternalError(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetCurrentIfStaleDueToInternalError() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != InternalError {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetStaleDueToInvalid(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetMaybeStaleDueToInvalid(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetCurrentIfStaleDueToInvalid() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != Invalid {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetStaleDueToRepositoryAdded(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = RepositoryAdded
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryAdded,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *EnvironmentStatus) SetMaybeStaleDueToRepositoryAdded(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = RepositoryAdded
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryAdded,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *EnvironmentStatus) SetCurrentIfStaleDueToRepositoryAdded() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != RepositoryAdded {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetStaleDueToRepositoryMissingDefaultBranch(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = RepositoryMissingDefaultBranch
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryMissingDefaultBranch,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *EnvironmentStatus) SetMaybeStaleDueToRepositoryMissingDefaultBranch(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = RepositoryMissingDefaultBranch
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryMissingDefaultBranch,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *EnvironmentStatus) SetCurrentIfStaleDueToRepositoryMissingDefaultBranch() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != RepositoryMissingDefaultBranch {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetStaleDueToRepositoryNotAccessible(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetMaybeStaleDueToRepositoryNotAccessible(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetCurrentIfStaleDueToRepositoryNotAccessible() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != RepositoryNotAccessible {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetStaleDueToRepositoryNotApplicable(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = RepositoryNotApplicable
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryNotApplicable,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *EnvironmentStatus) SetMaybeStaleDueToRepositoryNotApplicable(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = RepositoryNotApplicable
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryNotApplicable,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *EnvironmentStatus) SetCurrentIfStaleDueToRepositoryNotApplicable() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != RepositoryNotApplicable {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetStaleDueToRepositoryNotFound(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetMaybeStaleDueToRepositoryNotFound(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetCurrentIfStaleDueToRepositoryNotFound() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != RepositoryNotFound {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetStaleDueToRepositoryNotSupported(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = RepositoryNotSupported
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryNotSupported,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *EnvironmentStatus) SetMaybeStaleDueToRepositoryNotSupported(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = RepositoryNotSupported
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryNotSupported,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *EnvironmentStatus) SetCurrentIfStaleDueToRepositoryNotSupported() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != RepositoryNotSupported {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetCurrent() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) IsCurrent() bool {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Status == v1.ConditionTrue
		}
	}
	return true
}

func (s *EnvironmentStatus) IsStale() bool {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *EnvironmentStatus) GetStaleCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *EnvironmentStatus) GetStaleReason() string {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Reason
		}
	}
	return ""
}

func (s *EnvironmentStatus) GetStaleStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *EnvironmentStatus) GetStaleMessage() string {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Message
		}
	}
	return ""
}

func (s *EnvironmentStatus) SetInvalidDueToAddFinalizerFailed(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetMaybeInvalidDueToAddFinalizerFailed(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetValidIfInvalidDueToAddFinalizerFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != AddFinalizerFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetInvalidDueToControllerMissing(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetMaybeInvalidDueToControllerMissing(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetValidIfInvalidDueToControllerMissing() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != ControllerMissing {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetInvalidDueToDeploymentNotFound(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = DeploymentNotFound
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  DeploymentNotFound,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *EnvironmentStatus) SetMaybeInvalidDueToDeploymentNotFound(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = DeploymentNotFound
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  DeploymentNotFound,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *EnvironmentStatus) SetValidIfInvalidDueToDeploymentNotFound() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != DeploymentNotFound {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetInvalidDueToFailedGettingOwnedObjects(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetMaybeInvalidDueToFailedGettingOwnedObjects(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetValidIfInvalidDueToFailedGettingOwnedObjects() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != FailedGettingOwnedObjects {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetInvalidDueToFinalizationFailed(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetMaybeInvalidDueToFinalizationFailed(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetValidIfInvalidDueToFinalizationFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != FinalizationFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetInvalidDueToInternalError(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetMaybeInvalidDueToInternalError(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetValidIfInvalidDueToInternalError() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != InternalError {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetInvalidDueToRepositoryNotSupported(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetMaybeInvalidDueToRepositoryNotSupported(message string, args ...interface{}) {
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

func (s *EnvironmentStatus) SetValidIfInvalidDueToRepositoryNotSupported() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != RepositoryNotSupported {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) SetValid() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *EnvironmentStatus) IsValid() bool {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Status == v1.ConditionTrue
		}
	}
	return true
}

func (s *EnvironmentStatus) IsInvalid() bool {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *EnvironmentStatus) GetInvalidCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *EnvironmentStatus) GetInvalidReason() string {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Reason
		}
	}
	return ""
}

func (s *EnvironmentStatus) GetInvalidStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *EnvironmentStatus) GetInvalidMessage() string {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Message
		}
	}
	return ""
}

func (s *EnvironmentStatus) GetConditions() []v1.Condition {
	return s.Conditions
}

func (s *EnvironmentStatus) SetConditions(conditions []v1.Condition) {
	s.Conditions = conditions
}

func (s *EnvironmentStatus) ClearStaleConditions(currentGeneration int64) {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.ObservedGeneration >= currentGeneration {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (o *Environment) GetStatus() k8s.CommonStatus {
	return &o.Status
}
