package justest

import (
	"context"
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	"reflect"
)

type Extractor func(context.Context, JustT, any) (any, bool)

type ValueExtractor map[reflect.Kind]Extractor

func NewValueExtractor(defaultExtractor Extractor) ValueExtractor {
	ve := make(map[reflect.Kind]Extractor)
	ve[reflect.Invalid] = defaultExtractor
	return ve
}

func (ve ValueExtractor) MustExtractSingleValue(ctx context.Context, t JustT, actuals ...any) any {
	GetHelper(t).Helper()

	if len(actuals) != 1 {
		t.Fatalf("Expected exactly 1 actual value, found %d values: %+v", len(actuals), actuals)
	}

	value, found := ve.ExtractValue(ctx, t, actuals[0])
	if !found {
		t.Fatalf("Value not found for: %+v", actuals[0])
	}
	return value
}

func (ve ValueExtractor) ExtractValue(ctx context.Context, t JustT, actual any) (any, bool) {
	GetHelper(t).Helper()

	if actual == nil {
		return nil, true
	}

	v := reflect.ValueOf(actual)
	extractor, ok := ve[v.Kind()]
	if !ok {
		extractor, ok = ve[reflect.Invalid]
		if !ok {
			t.Fatalf("Default extractor is missing")
		}
	}

	return extractor(ctx, t, actual)
}

func (ve ValueExtractor) ExtractValues(ctx context.Context, t JustT, actuals ...any) []any {
	GetHelper(t).Helper()

	var resolved []any
	for _, actual := range actuals {
		value, found := ve.ExtractValue(ctx, t, actual)
		if found {
			resolved = append(resolved, value)
		}
	}
	return resolved
}

var (
	ExtractSameValue Extractor = func(_ context.Context, t JustT, v any) (any, bool) {
		GetHelper(t).Helper()
		return v, true
	}
	ExtractorValueLen Extractor = func(_ context.Context, t JustT, v any) (any, bool) {
		GetHelper(t).Helper()
		return reflect.ValueOf(v).Len(), true
	}
	ExtractorUnsupported Extractor = func(_ context.Context, t JustT, v any) (any, bool) {
		GetHelper(t).Helper()
		t.Fatalf("Unsupported actual value: %+v", v)
		panic("unreachable")
	}
)

func NewChannelExtractor(ve ValueExtractor, recurse bool) Extractor {
	return func(ctx context.Context, t JustT, v any) (any, bool) {
		GetHelper(t).Helper()

		msg, ok := reflect.ValueOf(v).TryRecv()
		if ok {
			if recurse {
				return ve.ExtractValue(ctx, t, msg.Interface())
			} else {
				return msg.Interface(), true
			}
		} else {
			return nil, false
		}
	}
}

func NewPointerExtractor(ve ValueExtractor, recurse bool) Extractor {
	return func(ctx context.Context, t JustT, v any) (any, bool) {
		GetHelper(t).Helper()

		underlyingValue := reflect.ValueOf(v).Elem()
		if recurse {
			return ve.ExtractValue(ctx, t, underlyingValue.Interface())
		} else {
			return underlyingValue.Interface(), true
		}
	}
}

func NewFuncExtractor(ve ValueExtractor, recurse bool) Extractor {
	return func(ctx context.Context, t JustT, v any) (any, bool) {
		GetHelper(t).Helper()

		funcValue := reflect.ValueOf(v)
		funcType := funcValue.Type()

		var in []reflect.Value
		switch funcType.NumIn() {
		case 0:
			in = nil
		case 1:
			arg0Type := funcType.In(0)
			if arg0Type.PkgPath() == justTTypePkgPath && arg0Type.Name() == justTTypeName {
				in = append(in, reflect.ValueOf(t))
			} else {
				t.Fatalf("Argument of functions with one argument must be of type JustT, found: %+v", arg0Type.Name())
				panic("unreachable")
			}
		default:
			t.Fatalf("Functions with more than 1 input parameter are not supported in this context: %+v", v)
			panic("unreachable")
		}

		switch funcType.NumOut() {
		case 0:
			funcValue.Call(in)
			return nil, false
		case 1:
			returnValues := funcValue.Call(in)
			if lang.IsErrorType(funcType.Out(0)) {
				if returnValues[0].IsNil() {
					return nil, true
				} else {
					t.Fatalf("Function failed: %+v", returnValues[0].Interface())
					panic("unreachable")
				}
			} else if recurse {
				return ve.ExtractValue(ctx, t, returnValues[0].Interface())
			} else {
				return returnValues[0].Interface(), true
			}
		case 2:
			if lang.IsErrorType(funcType.Out(1)) {
				returnValues := funcValue.Call(in)
				if returnValues[1].IsNil() {
					if recurse {
						return ve.ExtractValue(ctx, t, returnValues[0].Interface())
					} else {
						return returnValues[0].Interface(), true
					}
				} else {
					t.Fatalf("Function failed: %+v", returnValues[1].Interface())
					panic("unreachable")
				}
			} else {
				t.Fatalf("Functions with 2 return values must return 'error' as the 2nd return value: %+v", v)
				panic("unreachable")
			}
		default:
			t.Fatalf("Functions with %d return values are not supported: %+v", funcType.NumOut(), v)
			panic("unreachable")
		}
	}
}
