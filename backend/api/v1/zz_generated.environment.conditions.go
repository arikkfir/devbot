//go:build !ignore_autogenerated

// Code generated by devbot script. DO NOT EDIT.

package v1

import (
	. "github.com/arikkfir/devbot/backend/internal/util/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *EnvironmentStatus) GetCondition(conditionType string) *v1.Condition {
	return GetCondition(s.Conditions, conditionType)
}

func (s *EnvironmentStatus) SetFailedToInitializeDueToInternalError(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetMaybeFailedToInitializeDueToInternalError(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetInitializedIfFailedToInitializeDueToAnyOf(reasons ...string) bool {
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

func (s *EnvironmentStatus) SetInitialized() bool {
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

func (s *EnvironmentStatus) IsInitialized() bool {
	return !HasCondition(s.Conditions, FailedToInitialize) || IsConditionStatusOneOf(s.Conditions, FailedToInitialize, v1.ConditionFalse)
}

func (s *EnvironmentStatus) IsFailedToInitialize() bool {
	return IsConditionStatusOneOf(s.Conditions, FailedToInitialize, v1.ConditionTrue, v1.ConditionUnknown)
}

func (s *EnvironmentStatus) GetFailedToInitializeCondition() *v1.Condition {
	return GetCondition(s.Conditions, FailedToInitialize)
}

func (s *EnvironmentStatus) GetFailedToInitializeReason() string {
	return GetConditionReason(s.Conditions, FailedToInitialize)
}

func (s *EnvironmentStatus) GetFailedToInitializeStatus() *v1.ConditionStatus {
	return GetConditionStatus(s.Conditions, FailedToInitialize)
}

func (s *EnvironmentStatus) GetFailedToInitializeMessage() string {
	return GetConditionMessage(s.Conditions, FailedToInitialize)
}

func (s *EnvironmentStatus) SetFinalizingDueToFinalizationFailed(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetMaybeFinalizingDueToFinalizationFailed(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetFinalizingDueToFinalizerRemovalFailed(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetMaybeFinalizingDueToFinalizerRemovalFailed(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetFinalizingDueToInProgress(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetMaybeFinalizingDueToInProgress(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetFinalizedIfFinalizingDueToAnyOf(reasons ...string) bool {
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

func (s *EnvironmentStatus) SetFinalized() bool {
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

func (s *EnvironmentStatus) IsFinalized() bool {
	return !HasCondition(s.Conditions, Finalizing) || IsConditionStatusOneOf(s.Conditions, Finalizing, v1.ConditionFalse)
}

func (s *EnvironmentStatus) IsFinalizing() bool {
	return IsConditionStatusOneOf(s.Conditions, Finalizing, v1.ConditionTrue, v1.ConditionUnknown)
}

func (s *EnvironmentStatus) GetFinalizingCondition() *v1.Condition {
	return GetCondition(s.Conditions, Finalizing)
}

func (s *EnvironmentStatus) GetFinalizingReason() string {
	return GetConditionReason(s.Conditions, Finalizing)
}

func (s *EnvironmentStatus) GetFinalizingStatus() *v1.ConditionStatus {
	return GetConditionStatus(s.Conditions, Finalizing)
}

func (s *EnvironmentStatus) GetFinalizingMessage() string {
	return GetConditionMessage(s.Conditions, Finalizing)
}

func (s *EnvironmentStatus) SetInvalidDueToControllerNotAccessible(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetMaybeInvalidDueToControllerNotAccessible(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetInvalidDueToControllerNotFound(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetMaybeInvalidDueToControllerNotFound(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetInvalidDueToControllerReferenceMissing(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetMaybeInvalidDueToControllerReferenceMissing(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetInvalidDueToInternalError(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetMaybeInvalidDueToInternalError(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetValidIfInvalidDueToAnyOf(reasons ...string) bool {
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

func (s *EnvironmentStatus) SetValid() bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Valid]; !ok || v != "Yes" {
		s.PrivateArea[Valid] = "Yes"
		changed = true
	}
	changed = RemoveConditionIfReasonIsOneOf(&s.Conditions, Invalid, ControllerNotAccessible, ControllerNotFound, ControllerReferenceMissing, InternalError, "NonExistent") || changed
	return changed
}

func (s *EnvironmentStatus) IsValid() bool {
	return !HasCondition(s.Conditions, Invalid) || IsConditionStatusOneOf(s.Conditions, Invalid, v1.ConditionFalse)
}

func (s *EnvironmentStatus) IsInvalid() bool {
	return IsConditionStatusOneOf(s.Conditions, Invalid, v1.ConditionTrue, v1.ConditionUnknown)
}

func (s *EnvironmentStatus) GetInvalidCondition() *v1.Condition {
	return GetCondition(s.Conditions, Invalid)
}

func (s *EnvironmentStatus) GetInvalidReason() string {
	return GetConditionReason(s.Conditions, Invalid)
}

func (s *EnvironmentStatus) GetInvalidStatus() *v1.ConditionStatus {
	return GetConditionStatus(s.Conditions, Invalid)
}

func (s *EnvironmentStatus) GetInvalidMessage() string {
	return GetConditionMessage(s.Conditions, Invalid)
}

func (s *EnvironmentStatus) SetStaleDueToDeploymentsAreStale(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+DeploymentsAreStale {
		s.PrivateArea[Current] = "No: " + DeploymentsAreStale
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionTrue, DeploymentsAreStale, message, args...) || changed
	return changed
}

func (s *EnvironmentStatus) SetMaybeStaleDueToDeploymentsAreStale(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+DeploymentsAreStale {
		s.PrivateArea[Current] = "No: " + DeploymentsAreStale
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionUnknown, DeploymentsAreStale, message, args...) || changed
	return changed
}

func (s *EnvironmentStatus) SetStaleDueToFailedCreatingDeployment(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+FailedCreatingDeployment {
		s.PrivateArea[Current] = "No: " + FailedCreatingDeployment
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionTrue, FailedCreatingDeployment, message, args...) || changed
	return changed
}

func (s *EnvironmentStatus) SetMaybeStaleDueToFailedCreatingDeployment(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+FailedCreatingDeployment {
		s.PrivateArea[Current] = "No: " + FailedCreatingDeployment
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionUnknown, FailedCreatingDeployment, message, args...) || changed
	return changed
}

func (s *EnvironmentStatus) SetStaleDueToFailedDeletingDeployment(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+FailedDeletingDeployment {
		s.PrivateArea[Current] = "No: " + FailedDeletingDeployment
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionTrue, FailedDeletingDeployment, message, args...) || changed
	return changed
}

func (s *EnvironmentStatus) SetMaybeStaleDueToFailedDeletingDeployment(message string, args ...interface{}) bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "No: "+FailedDeletingDeployment {
		s.PrivateArea[Current] = "No: " + FailedDeletingDeployment
		changed = true
	}
	changed = SetCondition(&s.Conditions, Stale, v1.ConditionUnknown, FailedDeletingDeployment, message, args...) || changed
	return changed
}

func (s *EnvironmentStatus) SetStaleDueToInternalError(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetMaybeStaleDueToInternalError(message string, args ...interface{}) bool {
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

func (s *EnvironmentStatus) SetCurrentIfStaleDueToAnyOf(reasons ...string) bool {
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

func (s *EnvironmentStatus) SetCurrent() bool {
	changed := false
	if s.PrivateArea == nil {
		s.PrivateArea = make(map[string]string)
	}
	if v, ok := s.PrivateArea[Current]; !ok || v != "Yes" {
		s.PrivateArea[Current] = "Yes"
		changed = true
	}
	changed = RemoveConditionIfReasonIsOneOf(&s.Conditions, Stale, DeploymentsAreStale, FailedCreatingDeployment, FailedDeletingDeployment, InternalError, "NonExistent") || changed
	return changed
}

func (s *EnvironmentStatus) IsCurrent() bool {
	return !HasCondition(s.Conditions, Stale) || IsConditionStatusOneOf(s.Conditions, Stale, v1.ConditionFalse)
}

func (s *EnvironmentStatus) IsStale() bool {
	return IsConditionStatusOneOf(s.Conditions, Stale, v1.ConditionTrue, v1.ConditionUnknown)
}

func (s *EnvironmentStatus) GetStaleCondition() *v1.Condition {
	return GetCondition(s.Conditions, Stale)
}

func (s *EnvironmentStatus) GetStaleReason() string {
	return GetConditionReason(s.Conditions, Stale)
}

func (s *EnvironmentStatus) GetStaleStatus() *v1.ConditionStatus {
	return GetConditionStatus(s.Conditions, Stale)
}

func (s *EnvironmentStatus) GetStaleMessage() string {
	return GetConditionMessage(s.Conditions, Stale)
}

func (s *EnvironmentStatus) GetConditions() []v1.Condition {
	return s.Conditions
}

func (s *EnvironmentStatus) SetGenerationAndTransitionTime(generation int64) {
	SetConditionsGenerationAndTransitionTime(s.Conditions, generation)
}

func (s *EnvironmentStatus) ClearStaleConditions(currentGeneration int64) {
	ClearStaleConditions(&s.Conditions, currentGeneration)
}
