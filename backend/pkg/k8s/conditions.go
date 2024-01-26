package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type CommonStatusProvider interface {
	DeepCopyObject() runtime.Object
	GetStatus() CommonStatus
}

type CommonStatus interface {
	ClearStaleConditions(currentGeneration int64)
	GetConditions() []metav1.Condition
	SetConditions(conditions []metav1.Condition)
	SetInvalidDueToAddFinalizerFailed(message string, args ...interface{})
	SetInvalidDueToControllerMissing(message string, args ...interface{})
	SetInvalidDueToFailedGettingOwnedObjects(message string, args ...interface{})
	SetInvalidDueToFinalizationFailed(message string, args ...interface{})
	SetInvalidDueToInternalError(message string, args ...interface{})
	SetMaybeInvalidDueToAddFinalizerFailed(message string, args ...interface{})
	SetMaybeInvalidDueToControllerMissing(message string, args ...interface{})
	SetMaybeInvalidDueToFailedGettingOwnedObjects(message string, args ...interface{})
	SetMaybeInvalidDueToFinalizationFailed(message string, args ...interface{})
	SetMaybeInvalidDueToInternalError(message string, args ...interface{})
}
