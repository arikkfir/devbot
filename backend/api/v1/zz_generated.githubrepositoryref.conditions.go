package v1

import (
	"fmt"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *GitHubRepositoryRefStatus) SetStaleDueToCommitSHAOutOfSync(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = CommitSHAOutOfSync
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  CommitSHAOutOfSync,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryRefStatus) SetMaybeStaleDueToCommitSHAOutOfSync(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = CommitSHAOutOfSync
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  CommitSHAOutOfSync,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryRefStatus) SetCurrentIfStaleDueToCommitSHAOutOfSync() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != CommitSHAOutOfSync {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryRefStatus) SetStaleDueToRepositoryNameOutOfSync(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = RepositoryNameOutOfSync
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryNameOutOfSync,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryRefStatus) SetMaybeStaleDueToRepositoryNameOutOfSync(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = RepositoryNameOutOfSync
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryNameOutOfSync,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryRefStatus) SetCurrentIfStaleDueToRepositoryNameOutOfSync() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != RepositoryNameOutOfSync {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryRefStatus) SetStaleDueToRepositoryOwnerOutOfSync(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = RepositoryOwnerOutOfSync
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryOwnerOutOfSync,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryRefStatus) SetMaybeStaleDueToRepositoryOwnerOutOfSync(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = RepositoryOwnerOutOfSync
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryOwnerOutOfSync,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryRefStatus) SetCurrentIfStaleDueToRepositoryOwnerOutOfSync() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != RepositoryOwnerOutOfSync {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryRefStatus) SetCurrent() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryRefStatus) IsCurrent() bool {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Status == v1.ConditionTrue
		}
	}
	return true
}

func (s *GitHubRepositoryRefStatus) IsStale() bool {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *GitHubRepositoryRefStatus) GetStaleCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *GitHubRepositoryRefStatus) GetStaleReason() string {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Reason
		}
	}
	return ""
}

func (s *GitHubRepositoryRefStatus) GetStaleStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *GitHubRepositoryRefStatus) GetStaleMessage() string {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Message
		}
	}
	return ""
}

func (s *GitHubRepositoryRefStatus) SetInvalidDueToAddFinalizerFailed(message string, args ...interface{}) {
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

func (s *GitHubRepositoryRefStatus) SetMaybeInvalidDueToAddFinalizerFailed(message string, args ...interface{}) {
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

func (s *GitHubRepositoryRefStatus) SetValidIfInvalidDueToAddFinalizerFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != AddFinalizerFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryRefStatus) SetInvalidDueToControllerMissing(message string, args ...interface{}) {
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

func (s *GitHubRepositoryRefStatus) SetMaybeInvalidDueToControllerMissing(message string, args ...interface{}) {
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

func (s *GitHubRepositoryRefStatus) SetValidIfInvalidDueToControllerMissing() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != ControllerMissing {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryRefStatus) SetInvalidDueToFailedGettingOwnedObjects(message string, args ...interface{}) {
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

func (s *GitHubRepositoryRefStatus) SetMaybeInvalidDueToFailedGettingOwnedObjects(message string, args ...interface{}) {
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

func (s *GitHubRepositoryRefStatus) SetValidIfInvalidDueToFailedGettingOwnedObjects() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != FailedGettingOwnedObjects {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryRefStatus) SetInvalidDueToFinalizationFailed(message string, args ...interface{}) {
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

func (s *GitHubRepositoryRefStatus) SetMaybeInvalidDueToFinalizationFailed(message string, args ...interface{}) {
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

func (s *GitHubRepositoryRefStatus) SetValidIfInvalidDueToFinalizationFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != FinalizationFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryRefStatus) SetInvalidDueToInternalError(message string, args ...interface{}) {
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

func (s *GitHubRepositoryRefStatus) SetMaybeInvalidDueToInternalError(message string, args ...interface{}) {
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

func (s *GitHubRepositoryRefStatus) SetValidIfInvalidDueToInternalError() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != InternalError {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryRefStatus) SetValid() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryRefStatus) IsValid() bool {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Status == v1.ConditionTrue
		}
	}
	return true
}

func (s *GitHubRepositoryRefStatus) IsInvalid() bool {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *GitHubRepositoryRefStatus) GetInvalidCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *GitHubRepositoryRefStatus) GetInvalidReason() string {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Reason
		}
	}
	return ""
}

func (s *GitHubRepositoryRefStatus) GetInvalidStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *GitHubRepositoryRefStatus) GetInvalidMessage() string {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Message
		}
	}
	return ""
}

func (s *GitHubRepositoryRefStatus) GetConditions() []v1.Condition {
	return s.Conditions
}

func (s *GitHubRepositoryRefStatus) SetConditions(conditions []v1.Condition) {
	s.Conditions = conditions
}

func (s *GitHubRepositoryRefStatus) ClearStaleConditions(currentGeneration int64) {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.ObservedGeneration >= currentGeneration {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (o *GitHubRepositoryRef) GetStatus() k8s.CommonStatus {
	return &o.Status
}
