package v1

import (
	"fmt"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *DeploymentStatus) SetStaleDueToApplyFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = ApplyFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  ApplyFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetMaybeStaleDueToApplyFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = ApplyFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  ApplyFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetCurrentIfStaleDueToApplyFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != ApplyFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetStaleDueToBakingFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = BakingFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  BakingFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetMaybeStaleDueToBakingFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = BakingFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  BakingFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetCurrentIfStaleDueToBakingFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != BakingFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetStaleDueToCloneFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = CloneFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  CloneFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetMaybeStaleDueToCloneFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = CloneFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  CloneFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetCurrentIfStaleDueToCloneFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != CloneFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetStaleDueToCloneMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = CloneMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  CloneMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetMaybeStaleDueToCloneMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = CloneMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  CloneMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetCurrentIfStaleDueToCloneMissing() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != CloneMissing {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetStaleDueToCloning(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = Cloning
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  Cloning,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetMaybeStaleDueToCloning(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = Cloning
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  Cloning,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetCurrentIfStaleDueToCloning() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != Cloning {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetStaleDueToInternalError(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetMaybeStaleDueToInternalError(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetCurrentIfStaleDueToInternalError() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != InternalError {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetStaleDueToInvalid(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetMaybeStaleDueToInvalid(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetCurrentIfStaleDueToInvalid() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != Invalid {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetCurrent() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) IsCurrent() bool {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Status == v1.ConditionTrue
		}
	}
	return true
}

func (s *DeploymentStatus) IsStale() bool {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *DeploymentStatus) GetStaleCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *DeploymentStatus) GetStaleReason() string {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Reason
		}
	}
	return ""
}

func (s *DeploymentStatus) GetStaleStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *DeploymentStatus) GetStaleMessage() string {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Message
		}
	}
	return ""
}

func (s *DeploymentStatus) SetInvalidDueToAddFinalizerFailed(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetMaybeInvalidDueToAddFinalizerFailed(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetValidIfInvalidDueToAddFinalizerFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != AddFinalizerFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetInvalidDueToControllerMissing(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetMaybeInvalidDueToControllerMissing(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetValidIfInvalidDueToControllerMissing() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != ControllerMissing {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetInvalidDueToFailedGettingOwnedObjects(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetMaybeInvalidDueToFailedGettingOwnedObjects(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetValidIfInvalidDueToFailedGettingOwnedObjects() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != FailedGettingOwnedObjects {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetInvalidDueToFinalizationFailed(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetMaybeInvalidDueToFinalizationFailed(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetValidIfInvalidDueToFinalizationFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != FinalizationFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetInvalidDueToInternalError(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetMaybeInvalidDueToInternalError(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetValidIfInvalidDueToInternalError() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != InternalError {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetInvalidDueToRepositoryNotAccessible(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = RepositoryNotAccessible
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryNotAccessible,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetMaybeInvalidDueToRepositoryNotAccessible(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = RepositoryNotAccessible
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryNotAccessible,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetValidIfInvalidDueToRepositoryNotAccessible() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != RepositoryNotAccessible {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetInvalidDueToRepositoryNotFound(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = RepositoryNotFound
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryNotFound,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetMaybeInvalidDueToRepositoryNotFound(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = RepositoryNotFound
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryNotFound,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *DeploymentStatus) SetValidIfInvalidDueToRepositoryNotFound() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != RepositoryNotFound {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetInvalidDueToRepositoryNotSupported(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetMaybeInvalidDueToRepositoryNotSupported(message string, args ...interface{}) {
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

func (s *DeploymentStatus) SetValidIfInvalidDueToRepositoryNotSupported() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != RepositoryNotSupported {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) SetValid() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) IsValid() bool {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Status == v1.ConditionTrue
		}
	}
	return true
}

func (s *DeploymentStatus) IsInvalid() bool {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *DeploymentStatus) GetInvalidCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *DeploymentStatus) GetInvalidReason() string {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Reason
		}
	}
	return ""
}

func (s *DeploymentStatus) GetInvalidStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *DeploymentStatus) GetInvalidMessage() string {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Message
		}
	}
	return ""
}

func (s *DeploymentStatus) GetConditions() []v1.Condition {
	return s.Conditions
}

func (s *DeploymentStatus) SetConditions(conditions []v1.Condition) {
	s.Conditions = conditions
}

func (s *DeploymentStatus) ClearStaleConditions(currentGeneration int64) {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.ObservedGeneration >= currentGeneration {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (o *Deployment) GetStatus() k8s.CommonStatus {
	return &o.Status
}
