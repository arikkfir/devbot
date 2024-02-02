package k8s

import (
	"github.com/secureworks/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const statusFieldName = "Status"

func GetStatusOfType[T any](o client.Object) (T, bool) {
	// Get type of given object; if it's a pointer (normal, e.g. "client.Object") then get the type of the target
	typeOfValue := reflect.TypeOf(o)
	if typeOfValue.Kind() == reflect.Ptr {
		// Get the "Status" field of the object, if it exists
		if statusField, found := typeOfValue.Elem().FieldByName(statusFieldName); found {

			// Get the type of the desired interface
			var zero [0]T
			typeOfDesiredInterface := reflect.TypeOf(zero).Elem()

			// Verify that the "Status" field is indeed a struct
			if statusField.Type.Kind() == reflect.Struct {

				// We have to create a pointer to the status struct, since we need to return an interface, which is a pointer
				typeOfPtrToStatus := reflect.PointerTo(statusField.Type)
				if typeOfPtrToStatus.AssignableTo(typeOfDesiredInterface) {
					statusFieldValue := reflect.ValueOf(o).Elem().FieldByName(statusFieldName)
					statusValue := statusFieldValue.Addr()
					return statusValue.Interface().(T), true
				}
			}
		}
	}

	var zero [1]T
	return zero[0], false
}

func MustGetStatusOfType[T any](o client.Object) T {
	if status, found := GetStatusOfType[T](o); found {
		return status
	} else {
		var zero [0]T
		typeOfDesiredInterface := reflect.TypeOf(zero).Elem()
		panic(errors.New("type '%T' does not implement '%s'", o, typeOfDesiredInterface.Name()))
	}
}

type ResponseFunc func() *Result

type ErrorStrategy struct {
	OnSuccess         ResponseFunc
	OnNotFound        ResponseFunc
	OnConflict        ResponseFunc
	OnUnexpectedError func(err error) *Result
}

func WithStrategy(onSuccess ResponseFunc) *ErrorStrategy {
	return &ErrorStrategy{
		OnSuccess:  onSuccess,
		OnNotFound: DoNotRequeue,
		OnConflict: Requeue,
		OnUnexpectedError: func(err error) *Result {
			return RequeueDueToError(err)
		},
	}
}

func (s *ErrorStrategy) WithNotFound(onNotFound ResponseFunc) *ErrorStrategy {
	s.OnNotFound = onNotFound
	return s
}

func (s *ErrorStrategy) WithConflict(onConflict ResponseFunc) *ErrorStrategy {
	s.OnConflict = onConflict
	return s
}

func (s *ErrorStrategy) WithUnexpectedError(onUnexpectedError func(err error) *Result) *ErrorStrategy {
	s.OnUnexpectedError = onUnexpectedError
	return s
}

func RequeueAfterStrategy(interval time.Duration) ResponseFunc {
	return func() *Result { return RequeueAfter(interval) }
}

func RequeueDueToErrorStrategy(err error) ResponseFunc {
	return func() *Result { return RequeueDueToError(err) }
}

type ConditionsProvider interface {
	GetConditions() []metav1.Condition
	SetConditions(conditions []metav1.Condition)
}
