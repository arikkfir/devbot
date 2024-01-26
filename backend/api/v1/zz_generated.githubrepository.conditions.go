package v1

import (
	"fmt"
	"github.com/arikkfir/devbot/backend/pkg/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *GitHubRepositoryStatus) SetUnauthenticatedDueToAuthSecretForbidden(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionTrue
			c.Reason = AuthSecretForbidden
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionTrue,
		Reason:  AuthSecretForbidden,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeUnauthenticatedDueToAuthSecretForbidden(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionUnknown
			c.Reason = AuthSecretForbidden
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionUnknown,
		Reason:  AuthSecretForbidden,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetAuthenticatedIfUnauthenticatedDueToAuthSecretForbidden() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Unauthenticated || c.Reason != AuthSecretForbidden {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetUnauthenticatedDueToAuthSecretGetFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionTrue
			c.Reason = AuthSecretGetFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionTrue,
		Reason:  AuthSecretGetFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeUnauthenticatedDueToAuthSecretGetFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionUnknown
			c.Reason = AuthSecretGetFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionUnknown,
		Reason:  AuthSecretGetFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetAuthenticatedIfUnauthenticatedDueToAuthSecretGetFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Unauthenticated || c.Reason != AuthSecretGetFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetUnauthenticatedDueToAuthSecretKeyNotFound(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionTrue
			c.Reason = AuthSecretKeyNotFound
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionTrue,
		Reason:  AuthSecretKeyNotFound,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeUnauthenticatedDueToAuthSecretKeyNotFound(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionUnknown
			c.Reason = AuthSecretKeyNotFound
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionUnknown,
		Reason:  AuthSecretKeyNotFound,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetAuthenticatedIfUnauthenticatedDueToAuthSecretKeyNotFound() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Unauthenticated || c.Reason != AuthSecretKeyNotFound {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetUnauthenticatedDueToAuthSecretNotFound(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionTrue
			c.Reason = AuthSecretNotFound
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionTrue,
		Reason:  AuthSecretNotFound,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeUnauthenticatedDueToAuthSecretNotFound(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionUnknown
			c.Reason = AuthSecretNotFound
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionUnknown,
		Reason:  AuthSecretNotFound,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetAuthenticatedIfUnauthenticatedDueToAuthSecretNotFound() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Unauthenticated || c.Reason != AuthSecretNotFound {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetUnauthenticatedDueToAuthTokenEmpty(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionTrue
			c.Reason = AuthTokenEmpty
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionTrue,
		Reason:  AuthTokenEmpty,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeUnauthenticatedDueToAuthTokenEmpty(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionUnknown
			c.Reason = AuthTokenEmpty
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionUnknown,
		Reason:  AuthTokenEmpty,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetAuthenticatedIfUnauthenticatedDueToAuthTokenEmpty() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Unauthenticated || c.Reason != AuthTokenEmpty {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetUnauthenticatedDueToInvalid(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionTrue
			c.Reason = Invalid
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionTrue,
		Reason:  Invalid,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeUnauthenticatedDueToInvalid(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionUnknown
			c.Reason = Invalid
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionUnknown,
		Reason:  Invalid,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetAuthenticatedIfUnauthenticatedDueToInvalid() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Unauthenticated || c.Reason != Invalid {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetUnauthenticatedDueToTokenValidationFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionTrue
			c.Reason = TokenValidationFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionTrue,
		Reason:  TokenValidationFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeUnauthenticatedDueToTokenValidationFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Unauthenticated {
			c.Status = v1.ConditionUnknown
			c.Reason = TokenValidationFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Unauthenticated,
		Status:  v1.ConditionUnknown,
		Reason:  TokenValidationFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetAuthenticatedIfUnauthenticatedDueToTokenValidationFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Unauthenticated || c.Reason != TokenValidationFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetAuthenticated() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Unauthenticated {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) IsAuthenticated() bool {
	for _, c := range s.Conditions {
		if c.Type == Unauthenticated {
			return c.Status == v1.ConditionTrue
		}
	}
	return true
}

func (s *GitHubRepositoryStatus) IsUnauthenticated() bool {
	for _, c := range s.Conditions {
		if c.Type == Unauthenticated {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *GitHubRepositoryStatus) GetUnauthenticatedCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == Unauthenticated {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *GitHubRepositoryStatus) GetUnauthenticatedReason() string {
	for _, c := range s.Conditions {
		if c.Type == Unauthenticated {
			return c.Reason
		}
	}
	return ""
}

func (s *GitHubRepositoryStatus) GetUnauthenticatedStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == Unauthenticated {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *GitHubRepositoryStatus) GetUnauthenticatedMessage() string {
	for _, c := range s.Conditions {
		if c.Type == Unauthenticated {
			return c.Message
		}
	}
	return ""
}

func (s *GitHubRepositoryStatus) SetStaleDueToBranchesOutOfSync(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = BranchesOutOfSync
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  BranchesOutOfSync,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeStaleDueToBranchesOutOfSync(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = BranchesOutOfSync
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  BranchesOutOfSync,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetCurrentIfStaleDueToBranchesOutOfSync() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != BranchesOutOfSync {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetStaleDueToDefaultBranchOutOfSync(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = DefaultBranchOutOfSync
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  DefaultBranchOutOfSync,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeStaleDueToDefaultBranchOutOfSync(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = DefaultBranchOutOfSync
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  DefaultBranchOutOfSync,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetCurrentIfStaleDueToDefaultBranchOutOfSync() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != DefaultBranchOutOfSync {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetStaleDueToGitHubAPIFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = GitHubAPIFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  GitHubAPIFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeStaleDueToGitHubAPIFailed(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = GitHubAPIFailed
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  GitHubAPIFailed,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetCurrentIfStaleDueToGitHubAPIFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != GitHubAPIFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetStaleDueToInternalError(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetMaybeStaleDueToInternalError(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetCurrentIfStaleDueToInternalError() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != InternalError {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetStaleDueToInvalid(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetMaybeStaleDueToInvalid(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetCurrentIfStaleDueToInvalid() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != Invalid {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetStaleDueToRepositoryNotFound(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetMaybeStaleDueToRepositoryNotFound(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetCurrentIfStaleDueToRepositoryNotFound() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != RepositoryNotFound {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetStaleDueToUnauthenticated(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionTrue
			c.Reason = Unauthenticated
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionTrue,
		Reason:  Unauthenticated,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeStaleDueToUnauthenticated(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Stale {
			c.Status = v1.ConditionUnknown
			c.Reason = Unauthenticated
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Stale,
		Status:  v1.ConditionUnknown,
		Reason:  Unauthenticated,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetCurrentIfStaleDueToUnauthenticated() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale || c.Reason != Unauthenticated {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetCurrent() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Stale {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) IsCurrent() bool {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Status == v1.ConditionTrue
		}
	}
	return true
}

func (s *GitHubRepositoryStatus) IsStale() bool {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *GitHubRepositoryStatus) GetStaleCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *GitHubRepositoryStatus) GetStaleReason() string {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Reason
		}
	}
	return ""
}

func (s *GitHubRepositoryStatus) GetStaleStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *GitHubRepositoryStatus) GetStaleMessage() string {
	for _, c := range s.Conditions {
		if c.Type == Stale {
			return c.Message
		}
	}
	return ""
}

func (s *GitHubRepositoryStatus) SetInvalidDueToAddFinalizerFailed(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetMaybeInvalidDueToAddFinalizerFailed(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetValidIfInvalidDueToAddFinalizerFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != AddFinalizerFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetInvalidDueToAuthConfigMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = AuthConfigMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  AuthConfigMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeInvalidDueToAuthConfigMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = AuthConfigMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  AuthConfigMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetValidIfInvalidDueToAuthConfigMissing() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != AuthConfigMissing {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetInvalidDueToAuthSecretKeyMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = AuthSecretKeyMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  AuthSecretKeyMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeInvalidDueToAuthSecretKeyMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = AuthSecretKeyMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  AuthSecretKeyMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetValidIfInvalidDueToAuthSecretKeyMissing() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != AuthSecretKeyMissing {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetInvalidDueToAuthSecretNameMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = AuthSecretNameMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  AuthSecretNameMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeInvalidDueToAuthSecretNameMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = AuthSecretNameMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  AuthSecretNameMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetValidIfInvalidDueToAuthSecretNameMissing() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != AuthSecretNameMissing {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetInvalidDueToControllerMissing(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetMaybeInvalidDueToControllerMissing(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetValidIfInvalidDueToControllerMissing() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != ControllerMissing {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetInvalidDueToFailedGettingOwnedObjects(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetMaybeInvalidDueToFailedGettingOwnedObjects(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetValidIfInvalidDueToFailedGettingOwnedObjects() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != FailedGettingOwnedObjects {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetInvalidDueToFinalizationFailed(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetMaybeInvalidDueToFinalizationFailed(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetValidIfInvalidDueToFinalizationFailed() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != FinalizationFailed {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetInvalidDueToInternalError(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetMaybeInvalidDueToInternalError(message string, args ...interface{}) {
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

func (s *GitHubRepositoryStatus) SetValidIfInvalidDueToInternalError() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != InternalError {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetInvalidDueToInvalidRefreshInterval(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = InvalidRefreshInterval
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  InvalidRefreshInterval,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeInvalidDueToInvalidRefreshInterval(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = InvalidRefreshInterval
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  InvalidRefreshInterval,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetValidIfInvalidDueToInvalidRefreshInterval() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != InvalidRefreshInterval {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetInvalidDueToRepositoryNameMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = RepositoryNameMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryNameMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeInvalidDueToRepositoryNameMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = RepositoryNameMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryNameMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetValidIfInvalidDueToRepositoryNameMissing() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != RepositoryNameMissing {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetInvalidDueToRepositoryOwnerMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionTrue
			c.Reason = RepositoryOwnerMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionTrue,
		Reason:  RepositoryOwnerMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetMaybeInvalidDueToRepositoryOwnerMissing(message string, args ...interface{}) {
	for i, c := range s.Conditions {
		if c.Type == Invalid {
			c.Status = v1.ConditionUnknown
			c.Reason = RepositoryOwnerMissing
			c.Message = fmt.Sprintf(message, args...)
			s.Conditions[i] = c
			return
		}
	}
	s.Conditions = append(s.Conditions, v1.Condition{
		Type:    Invalid,
		Status:  v1.ConditionUnknown,
		Reason:  RepositoryOwnerMissing,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *GitHubRepositoryStatus) SetValidIfInvalidDueToRepositoryOwnerMissing() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid || c.Reason != RepositoryOwnerMissing {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) SetValid() {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.Type != Invalid {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (s *GitHubRepositoryStatus) IsValid() bool {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Status == v1.ConditionTrue
		}
	}
	return true
}

func (s *GitHubRepositoryStatus) IsInvalid() bool {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Status == v1.ConditionTrue || c.Status == v1.ConditionUnknown
		}
	}
	return false
}

func (s *GitHubRepositoryStatus) GetInvalidCondition() *v1.Condition {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			lc := c
			return &lc
		}
	}
	return nil
}

func (s *GitHubRepositoryStatus) GetInvalidReason() string {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Reason
		}
	}
	return ""
}

func (s *GitHubRepositoryStatus) GetInvalidStatus() *v1.ConditionStatus {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			status := c.Status
			return &status
		}
	}
	return nil
}

func (s *GitHubRepositoryStatus) GetInvalidMessage() string {
	for _, c := range s.Conditions {
		if c.Type == Invalid {
			return c.Message
		}
	}
	return ""
}

func (s *GitHubRepositoryStatus) GetConditions() []v1.Condition {
	return s.Conditions
}

func (s *GitHubRepositoryStatus) SetConditions(conditions []v1.Condition) {
	s.Conditions = conditions
}

func (s *GitHubRepositoryStatus) ClearStaleConditions(currentGeneration int64) {
	var newConditions []v1.Condition
	for _, c := range s.Conditions {
		if c.ObservedGeneration >= currentGeneration {
			newConditions = append(newConditions, c)
		}
	}
	s.Conditions = newConditions
}

func (o *GitHubRepository) GetStatus() k8s.CommonStatus {
	return &o.Status
}
