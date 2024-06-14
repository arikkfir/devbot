//go:build !ignore_autogenerated

// Code generated by devbot script. DO NOT EDIT.

package v1

import (
	. "github.com/arikkfir/devbot/backend/internal/util/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *ApplicationStatus) GetCondition(conditionType string) *v1.Condition {
	return GetCondition(s.Conditions, conditionType)
}

func (s *ApplicationStatus) SetFailedToInitializeDueToInternalError(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Initialized]; !ok || v != "No: "+InternalError {
		s.PrivateArea[Initialized] = "No: " + InternalError
		changed = true
	}
	changed = SetCondition(&s.Conditions, FailedToInitialize, v1.ConditionTrue, InternalError, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetMaybeFailedToInitializeDueToInternalError(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Initialized]; !ok || v != "No: "+InternalError {
		s.PrivateArea[Initialized] = "No: " + InternalError
		changed = true
	}
	changed = SetCondition(&s.Conditions, FailedToInitialize, v1.ConditionUnknown, InternalError, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetInitializedIfFailedToInitializeDueToAnyOf(reasons ...string) bool {
	changed := false
	changed = RemoveConditionIfReasonIsOneOf(&s.Conditions, FailedToInitialize, reasons...) || changed
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if s.IsInitialized() {
		if v, ok := s.PrivateArea[Initialized]; !ok || v != "Yes" {
			s.PrivateArea[Initialized] = "Yes"
			changed = true
		}
	} else {
		if v, ok := s.PrivateArea[Initialized]; !ok || v != "No: "+s.GetFailedToInitializeReason() {
			s.PrivateArea[Initialized] = "No: " + s.GetFailedToInitializeReason()
			changed = true
		}
	}
	return changed
}

func (s *ApplicationStatus) SetInitialized() bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Initialized]; !ok || v != "Yes" {
		s.PrivateArea[Initialized] = "Yes"
		changed = true
	}
	changed = RemoveConditionIfReasonIsOneOf(&s.Conditions, FailedToInitialize, InternalError, "NonExistent") || changed
	return changed
}

func (s *ApplicationStatus) IsInitialized() bool {
	return !HasCondition(s.Conditions, FailedToInitialize) || IsConditionStatusOneOf(s.Conditions, FailedToInitialize, v1.ConditionFalse)
}

func (s *ApplicationStatus) IsFailedToInitialize() bool {
	return IsConditionStatusOneOf(s.Conditions, FailedToInitialize, v1.ConditionTrue, v1.ConditionUnknown)
}

func (s *ApplicationStatus) GetFailedToInitializeCondition() *v1.Condition {
	return GetCondition(s.Conditions, FailedToInitialize)
}

func (s *ApplicationStatus) GetFailedToInitializeReason() string {
	return GetConditionReason(s.Conditions, FailedToInitialize)
}

func (s *ApplicationStatus) GetFailedToInitializeStatus() *v1.ConditionStatus {
	return GetConditionStatus(s.Conditions, FailedToInitialize)
}

func (s *ApplicationStatus) GetFailedToInitializeMessage() string {
	return GetConditionMessage(s.Conditions, FailedToInitialize)
}

func (s *ApplicationStatus) SetFinalizingDueToFinalizationFailed(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Finalized]; !ok || v != "No: "+FinalizationFailed {
		s.PrivateArea[Finalized] = "No: " + FinalizationFailed
		changed = true
	}
	changed = SetCondition(&s.Conditions, Finalizing, v1.ConditionTrue, FinalizationFailed, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetMaybeFinalizingDueToFinalizationFailed(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Finalized]; !ok || v != "No: "+FinalizationFailed {
		s.PrivateArea[Finalized] = "No: " + FinalizationFailed
		changed = true
	}
	changed = SetCondition(&s.Conditions, Finalizing, v1.ConditionUnknown, FinalizationFailed, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetFinalizingDueToFinalizerRemovalFailed(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Finalized]; !ok || v != "No: "+FinalizerRemovalFailed {
		s.PrivateArea[Finalized] = "No: " + FinalizerRemovalFailed
		changed = true
	}
	changed = SetCondition(&s.Conditions, Finalizing, v1.ConditionTrue, FinalizerRemovalFailed, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetMaybeFinalizingDueToFinalizerRemovalFailed(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Finalized]; !ok || v != "No: "+FinalizerRemovalFailed {
		s.PrivateArea[Finalized] = "No: " + FinalizerRemovalFailed
		changed = true
	}
	changed = SetCondition(&s.Conditions, Finalizing, v1.ConditionUnknown, FinalizerRemovalFailed, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetFinalizingDueToInProgress(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Finalized]; !ok || v != "No: "+InProgress {
		s.PrivateArea[Finalized] = "No: " + InProgress
		changed = true
	}
	changed = SetCondition(&s.Conditions, Finalizing, v1.ConditionTrue, InProgress, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetMaybeFinalizingDueToInProgress(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Finalized]; !ok || v != "No: "+InProgress {
		s.PrivateArea[Finalized] = "No: " + InProgress
		changed = true
	}
	changed = SetCondition(&s.Conditions, Finalizing, v1.ConditionUnknown, InProgress, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetFinalizedIfFinalizingDueToAnyOf(reasons ...string) bool {
	changed := false
	changed = RemoveConditionIfReasonIsOneOf(&s.Conditions, Finalizing, reasons...) || changed
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if s.IsFinalized() {
		if v, ok := s.PrivateArea[Finalized]; !ok || v != "Yes" {
			s.PrivateArea[Finalized] = "Yes"
			changed = true
		}
	} else {
		if v, ok := s.PrivateArea[Finalized]; !ok || v != "No: "+s.GetFinalizingReason() {
			s.PrivateArea[Finalized] = "No: " + s.GetFinalizingReason()
			changed = true
		}
	}
	return changed
}

func (s *ApplicationStatus) SetFinalized() bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Finalized]; !ok || v != "Yes" {
		s.PrivateArea[Finalized] = "Yes"
		changed = true
	}
	changed = RemoveConditionIfReasonIsOneOf(&s.Conditions, Finalizing, FinalizationFailed, FinalizerRemovalFailed, InProgress, "NonExistent") || changed
	return changed
}

func (s *ApplicationStatus) IsFinalized() bool {
	return !HasCondition(s.Conditions, Finalizing) || IsConditionStatusOneOf(s.Conditions, Finalizing, v1.ConditionFalse)
}

func (s *ApplicationStatus) IsFinalizing() bool {
	return IsConditionStatusOneOf(s.Conditions, Finalizing, v1.ConditionTrue, v1.ConditionUnknown)
}

func (s *ApplicationStatus) GetFinalizingCondition() *v1.Condition {
	return GetCondition(s.Conditions, Finalizing)
}

func (s *ApplicationStatus) GetFinalizingReason() string {
	return GetConditionReason(s.Conditions, Finalizing)
}

func (s *ApplicationStatus) GetFinalizingStatus() *v1.ConditionStatus {
	return GetConditionStatus(s.Conditions, Finalizing)
}

func (s *ApplicationStatus) GetFinalizingMessage() string {
	return GetConditionMessage(s.Conditions, Finalizing)
}

func (s *ApplicationStatus) SetInvalidDueToControllerNotAccessible(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Valid]; !ok || v != "No: "+ControllerNotAccessible {
		s.PrivateArea[Valid] = "No: " + ControllerNotAccessible
		changed = true
	}
	changed = SetCondition(&s.Conditions, Invalid, v1.ConditionTrue, ControllerNotAccessible, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetMaybeInvalidDueToControllerNotAccessible(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Valid]; !ok || v != "No: "+ControllerNotAccessible {
		s.PrivateArea[Valid] = "No: " + ControllerNotAccessible
		changed = true
	}
	changed = SetCondition(&s.Conditions, Invalid, v1.ConditionUnknown, ControllerNotAccessible, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetInvalidDueToControllerNotFound(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Valid]; !ok || v != "No: "+ControllerNotFound {
		s.PrivateArea[Valid] = "No: " + ControllerNotFound
		changed = true
	}
	changed = SetCondition(&s.Conditions, Invalid, v1.ConditionTrue, ControllerNotFound, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetMaybeInvalidDueToControllerNotFound(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Valid]; !ok || v != "No: "+ControllerNotFound {
		s.PrivateArea[Valid] = "No: " + ControllerNotFound
		changed = true
	}
	changed = SetCondition(&s.Conditions, Invalid, v1.ConditionUnknown, ControllerNotFound, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetInvalidDueToControllerReferenceMissing(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Valid]; !ok || v != "No: "+ControllerReferenceMissing {
		s.PrivateArea[Valid] = "No: " + ControllerReferenceMissing
		changed = true
	}
	changed = SetCondition(&s.Conditions, Invalid, v1.ConditionTrue, ControllerReferenceMissing, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetMaybeInvalidDueToControllerReferenceMissing(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Valid]; !ok || v != "No: "+ControllerReferenceMissing {
		s.PrivateArea[Valid] = "No: " + ControllerReferenceMissing
		changed = true
	}
	changed = SetCondition(&s.Conditions, Invalid, v1.ConditionUnknown, ControllerReferenceMissing, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetInvalidDueToInternalError(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Valid]; !ok || v != "No: "+InternalError {
		s.PrivateArea[Valid] = "No: " + InternalError
		changed = true
	}
	changed = SetCondition(&s.Conditions, Invalid, v1.ConditionTrue, InternalError, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetMaybeInvalidDueToInternalError(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Valid]; !ok || v != "No: "+InternalError {
		s.PrivateArea[Valid] = "No: " + InternalError
		changed = true
	}
	changed = SetCondition(&s.Conditions, Invalid, v1.ConditionUnknown, InternalError, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetInvalidDueToInvalidBranchSpecification(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Valid]; !ok || v != "No: "+InvalidBranchSpecification {
		s.PrivateArea[Valid] = "No: " + InvalidBranchSpecification
		changed = true
	}
	changed = SetCondition(&s.Conditions, Invalid, v1.ConditionTrue, InvalidBranchSpecification, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetMaybeInvalidDueToInvalidBranchSpecification(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Valid]; !ok || v != "No: "+InvalidBranchSpecification {
		s.PrivateArea[Valid] = "No: " + InvalidBranchSpecification
		changed = true
	}
	changed = SetCondition(&s.Conditions, Invalid, v1.ConditionUnknown, InvalidBranchSpecification, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetValidIfInvalidDueToAnyOf(reasons ...string) bool {
	changed := false
	changed = RemoveConditionIfReasonIsOneOf(&s.Conditions, Invalid, reasons...) || changed
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if s.IsValid() {
		if v, ok := s.PrivateArea[Valid]; !ok || v != "Yes" {
			s.PrivateArea[Valid] = "Yes"
			changed = true
		}
	} else {
		if v, ok := s.PrivateArea[Valid]; !ok || v != "No: "+s.GetInvalidReason() {
			s.PrivateArea[Valid] = "No: " + s.GetInvalidReason()
			changed = true
		}
	}
	return changed
}

func (s *ApplicationStatus) SetValid() bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Valid]; !ok || v != "Yes" {
		s.PrivateArea[Valid] = "Yes"
		changed = true
	}
	changed = RemoveConditionIfReasonIsOneOf(&s.Conditions, Invalid, ControllerNotAccessible, ControllerNotFound, ControllerReferenceMissing, InternalError, InvalidBranchSpecification, "NonExistent") || changed
	return changed
}

func (s *ApplicationStatus) IsValid() bool {
	return !HasCondition(s.Conditions, Invalid) || IsConditionStatusOneOf(s.Conditions, Invalid, v1.ConditionFalse)
}

func (s *ApplicationStatus) IsInvalid() bool {
	return IsConditionStatusOneOf(s.Conditions, Invalid, v1.ConditionTrue, v1.ConditionUnknown)
}

func (s *ApplicationStatus) GetInvalidCondition() *v1.Condition {
	return GetCondition(s.Conditions, Invalid)
}

func (s *ApplicationStatus) GetInvalidReason() string {
	return GetConditionReason(s.Conditions, Invalid)
}

func (s *ApplicationStatus) GetInvalidStatus() *v1.ConditionStatus {
	return GetConditionStatus(s.Conditions, Invalid)
}

func (s *ApplicationStatus) GetInvalidMessage() string {
	return GetConditionMessage(s.Conditions, Invalid)
}

func (s *ApplicationStatus) SetStaleDueToEnvironmentsAreStale(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+EnvironmentsAreStale {
		s.PrivateArea[Current] = "No: " + EnvironmentsAreStale
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionTrue, EnvironmentsAreStale, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetMaybeStaleDueToEnvironmentsAreStale(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+EnvironmentsAreStale {
		s.PrivateArea[Current] = "No: " + EnvironmentsAreStale
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionUnknown, EnvironmentsAreStale, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetStaleDueToInternalError(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+InternalError {
		s.PrivateArea[Current] = "No: " + InternalError
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionTrue, InternalError, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetMaybeStaleDueToInternalError(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+InternalError {
		s.PrivateArea[Current] = "No: " + InternalError
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionUnknown, InternalError, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetStaleDueToRepositoryNotAccessible(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+RepositoryNotAccessible {
		s.PrivateArea[Current] = "No: " + RepositoryNotAccessible
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionTrue, RepositoryNotAccessible, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetMaybeStaleDueToRepositoryNotAccessible(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+RepositoryNotAccessible {
		s.PrivateArea[Current] = "No: " + RepositoryNotAccessible
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionUnknown, RepositoryNotAccessible, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetStaleDueToRepositoryNotFound(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+RepositoryNotFound {
		s.PrivateArea[Current] = "No: " + RepositoryNotFound
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionTrue, RepositoryNotFound, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetMaybeStaleDueToRepositoryNotFound(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+RepositoryNotFound {
		s.PrivateArea[Current] = "No: " + RepositoryNotFound
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionUnknown, RepositoryNotFound, message, args...) || changed
	return changed
}

func (s *ApplicationStatus) SetCurrentIfStaleDueToAnyOf(reasons ...string) bool {
	changed := false
	changed = RemoveConditionIfReasonIsOneOf(&s.Conditions, Stale, reasons...) || changed
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if s.IsCurrent() {
		if v, ok := s.PrivateArea[Current]; !ok || v != "Yes" {
			s.PrivateArea[Current] = "Yes"
			changed = true
		}
	} else {
		if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+s.GetStaleReason() {
			s.PrivateArea[Current] = "No: " + s.GetStaleReason()
			changed = true
		}
	}
	return changed
}

func (s *ApplicationStatus) SetCurrent() bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "Yes" {
		s.PrivateArea[Current] = "Yes"
		changed = true
	}
	changed = RemoveConditionIfReasonIsOneOf(&s.Conditions, Stale, EnvironmentsAreStale, InternalError, RepositoryNotAccessible, RepositoryNotFound, "NonExistent") || changed
	return changed
}

func (s *ApplicationStatus) IsCurrent() bool {
	return !HasCondition(s.Conditions, Stale) || IsConditionStatusOneOf(s.Conditions, Stale, v1.ConditionFalse)
}

func (s *ApplicationStatus) IsStale() bool {
	return IsConditionStatusOneOf(s.Conditions, Stale, v1.ConditionTrue, v1.ConditionUnknown)
}

func (s *ApplicationStatus) GetStaleCondition() *v1.Condition {
	return GetCondition(s.Conditions, Stale)
}

func (s *ApplicationStatus) GetStaleReason() string {
	return GetConditionReason(s.Conditions, Stale)
}

func (s *ApplicationStatus) GetStaleStatus() *v1.ConditionStatus {
	return GetConditionStatus(s.Conditions, Stale)
}

func (s *ApplicationStatus) GetStaleMessage() string {
	return GetConditionMessage(s.Conditions, Stale)
}

func (s *ApplicationStatus) GetConditions() []v1.Condition {
	return s.Conditions
}

func (s *ApplicationStatus) SetGenerationAndTransitionTime(generation int64) {
	SetConditionsGenerationAndTransitionTime(s.Conditions, generation)
}

func (s *ApplicationStatus) ClearStaleConditions(currentGeneration int64) {
	ClearStaleConditions(&s.Conditions, currentGeneration)
}
