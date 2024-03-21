package lang

import "reflect"

func IsErrorType(ot reflect.Type) bool {
	return ot.Kind() == reflect.Interface && ot.Implements(reflect.TypeOf((*error)(nil)).Elem())
}

func IsErrorValue(v reflect.Value) bool {
	return IsErrorType(v.Type())
}
