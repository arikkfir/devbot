package k8s

import (
	"fmt"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const StatusFieldName = "Status"

func GetStatusOfType[T any](o client.Object) (T, bool) {
	typeOfValue := reflect.TypeOf(o)
	if typeOfValue.Kind() == reflect.Ptr {
		if statusField, found := typeOfValue.Elem().FieldByName(StatusFieldName); found {
			var zero [0]T
			typeOfDesiredInterface := reflect.TypeOf(zero).Elem()
			if statusField.Type.Kind() == reflect.Struct {
				typeOfPtrToStatus := reflect.PointerTo(statusField.Type)
				if typeOfPtrToStatus.AssignableTo(typeOfDesiredInterface) {
					statusFieldValue := reflect.ValueOf(o).Elem().FieldByName(StatusFieldName)
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
		panic(fmt.Errorf("type '%T' does not implement '%s'", o, typeOfDesiredInterface.Name()))
	}
}
