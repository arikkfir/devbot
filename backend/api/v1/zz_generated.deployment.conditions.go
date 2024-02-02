//go:build !ignore_autogenerated

// Code generated by devbot script. DO NOT EDIT.

package v1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"slices"
)

func (s *DeploymentStatus) SetStaleDueToApplyFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != ApplyFailed || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = ApplyFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  ApplyFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeStaleDueToApplyFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != ApplyFailed || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = ApplyFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  ApplyFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetStaleDueToBakingFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != BakingFailed || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = BakingFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  BakingFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeStaleDueToBakingFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != BakingFailed || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = BakingFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  BakingFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetStaleDueToCheckoutFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != CheckoutFailed || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = CheckoutFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  CheckoutFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeStaleDueToCheckoutFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != CheckoutFailed || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = CheckoutFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  CheckoutFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetStaleDueToCloneFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != CloneFailed || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = CloneFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  CloneFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeStaleDueToCloneFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != CloneFailed || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = CloneFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  CloneFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetStaleDueToCloneMissing(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != CloneMissing || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = CloneMissing
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  CloneMissing,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeStaleDueToCloneMissing(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != CloneMissing || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = CloneMissing
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  CloneMissing,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetStaleDueToCloning(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != Cloning || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = Cloning
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  Cloning,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeStaleDueToCloning(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != Cloning || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = Cloning
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  Cloning,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetStaleDueToInternalError(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != InternalError || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = InternalError
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  InternalError,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeStaleDueToInternalError(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != InternalError || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = InternalError
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  InternalError,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetStaleDueToInvalid(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != Invalid || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = Invalid
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  Invalid,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeStaleDueToInvalid(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != Invalid || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = Invalid
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  Invalid,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetStaleDueToPullFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != PullFailed || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = PullFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  PullFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeStaleDueToPullFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != PullFailed || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = PullFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  PullFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetStaleDueToPulling(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != Pulling || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = Pulling
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  Pulling,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeStaleDueToPulling(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != Pulling || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = Pulling
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  Pulling,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetStaleDueToRepositoryNotAccessible(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != RepositoryNotAccessible || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = RepositoryNotAccessible
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryNotAccessible,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeStaleDueToRepositoryNotAccessible(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != RepositoryNotAccessible || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = RepositoryNotAccessible
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryNotAccessible,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetStaleDueToRepositoryNotFound(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != RepositoryNotFound || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = RepositoryNotFound
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryNotFound,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeStaleDueToRepositoryNotFound(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != RepositoryNotFound || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = RepositoryNotFound
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryNotFound,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetCurrentIfStaleDueToAnyOf(reasons ...string) bool {
	changed := false
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || !slices.Contains(reasons, c.Reason) {
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
			return c.Status != v1.ConditionTrue
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

func (s *DeploymentStatus) SetFinalizingDueToFinalizationFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Finalizing {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != FinalizationFailed || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = FinalizationFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Finalizing,
		Status:  v1.ConditionTrue,
		Reason:  FinalizationFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeFinalizingDueToFinalizationFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Finalizing {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != FinalizationFailed || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = FinalizationFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Finalizing,
		Status:  v1.ConditionUnknown,
		Reason:  FinalizationFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetFinalizingDueToFinalizerRemovalFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Finalizing {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != FinalizerRemovalFailed || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = FinalizerRemovalFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Finalizing,
		Status:  v1.ConditionTrue,
		Reason:  FinalizerRemovalFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeFinalizingDueToFinalizerRemovalFailed(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Finalizing {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != FinalizerRemovalFailed || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = FinalizerRemovalFailed
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Finalizing,
		Status:  v1.ConditionUnknown,
		Reason:  FinalizerRemovalFailed,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetFinalizingDueToInProgress(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Finalizing {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != InProgress || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = InProgress
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Finalizing,
		Status:  v1.ConditionTrue,
		Reason:  InProgress,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeFinalizingDueToInProgress(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Finalizing {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != InProgress || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = InProgress
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Finalizing,
		Status:  v1.ConditionUnknown,
		Reason:  InProgress,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetFinalizedIfFinalizingDueToAnyOf(reasons ...string) bool {
	changed := false
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Finalizing || !slices.Contains(reasons, c.Reason) {
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

func (s *DeploymentStatus) SetFinalized() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Finalizing {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) IsFinalized() bool {
	for _, c := range s.Conditions {
		if c.Type == Finalizing {
			return c.Status != v1.ConditionTrue
		}
	}
	return true
}

func (s *DeploymentStatus) IsFinalizing() bool {
	for _, c := range s.Conditions {
		if c.Type == Finalizing {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *DeploymentStatus) GetFinalizingCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == Finalizing {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *DeploymentStatus) GetFinalizingReason() string {
	for _, c := range s.Conditions {
		if c.Type == Finalizing {
			return c.Reason
		}
	}
	return ""
}

func (s *DeploymentStatus) GetFinalizingStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == Finalizing {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *DeploymentStatus) GetFinalizingMessage() string {
	for _, c := range s.Conditions {
		if c.Type == Finalizing {
			return c.Message
		}
	}
	return ""
}

func (s *DeploymentStatus) SetFailedToInitializeDueToInternalError(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == FailedToInitialize {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != InternalError || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = InternalError
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    FailedToInitialize,
		Status:  v1.ConditionTrue,
		Reason:  InternalError,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeFailedToInitializeDueToInternalError(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == FailedToInitialize {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != InternalError || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = InternalError
				c.Message = msg
				s.Conditions[i] = c
				return true
			} else {
				return false
			}
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    FailedToInitialize,
		Status:  v1.ConditionUnknown,
		Reason:  InternalError,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetInitializedIfFailedToInitializeDueToAnyOf(reasons ...string) bool {
	changed := false
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != FailedToInitialize || !slices.Contains(reasons, c.Reason) {
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

func (s *DeploymentStatus) SetInitialized() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != FailedToInitialize {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *DeploymentStatus) IsInitialized() bool {
	for _, c := range s.Conditions {
		if c.Type == FailedToInitialize {
			return c.Status != v1.ConditionTrue
		}
	}
	return true
}

func (s *DeploymentStatus) IsFailedToInitialize() bool {
	for _, c := range s.Conditions {
		if c.Type == FailedToInitialize {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *DeploymentStatus) GetFailedToInitializeCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == FailedToInitialize {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *DeploymentStatus) GetFailedToInitializeReason() string {
	for _, c := range s.Conditions {
		if c.Type == FailedToInitialize {
			return c.Reason
		}
	}
	return ""
}

func (s *DeploymentStatus) GetFailedToInitializeStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == FailedToInitialize {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *DeploymentStatus) GetFailedToInitializeMessage() string {
	for _, c := range s.Conditions {
		if c.Type == FailedToInitialize {
			return c.Message
		}
	}
	return ""
}

func (s *DeploymentStatus) SetInvalidDueToControllerNotAccessible(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != ControllerNotAccessible || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = ControllerNotAccessible
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
		Reason:  ControllerNotAccessible,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeInvalidDueToControllerNotAccessible(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != ControllerNotAccessible || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = ControllerNotAccessible
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
		Reason:  ControllerNotAccessible,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetInvalidDueToControllerNotFound(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != ControllerNotFound || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = ControllerNotFound
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
		Reason:  ControllerNotFound,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeInvalidDueToControllerNotFound(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != ControllerNotFound || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = ControllerNotFound
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
		Reason:  ControllerNotFound,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetInvalidDueToControllerReferenceMissing(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != ControllerReferenceMissing || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = ControllerReferenceMissing
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
		Reason:  ControllerReferenceMissing,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeInvalidDueToControllerReferenceMissing(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != ControllerReferenceMissing || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = ControllerReferenceMissing
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
		Reason:  ControllerReferenceMissing,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetInvalidDueToInternalError(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != InternalError || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = InternalError
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
		Reason:  InternalError,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeInvalidDueToInternalError(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != InternalError || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = InternalError
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
		Reason:  InternalError,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetInvalidDueToRepositoryNotSupported(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionTrue || c.Reason != RepositoryNotSupported || c.Message != msg {
				c.Status = v1.ConditionTrue
				c.Reason = RepositoryNotSupported
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
		Reason:  RepositoryNotSupported,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetMaybeInvalidDueToRepositoryNotSupported(message string, args ...interface{}) bool {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			msg := fmt.Sprintf(message, args...)
			if c.Status != v1.ConditionUnknown || c.Reason != RepositoryNotSupported || c.Message != msg {
				c.Status = v1.ConditionUnknown
				c.Reason = RepositoryNotSupported
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
		Reason:  RepositoryNotSupported,
		Message: fmt.Sprintf(message, args...),
	})
	return true
}

func (s *DeploymentStatus) SetValidIfInvalidDueToAnyOf(reasons ...string) bool {
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
			return c.Status != v1.ConditionTrue
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
