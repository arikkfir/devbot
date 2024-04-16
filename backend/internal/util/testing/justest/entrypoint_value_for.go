package justest

import (
	"fmt"
	"sync"
	"testing"
)

var (
	contextValues = sync.Map{}
)

func getValueForT(t T, key any) any {
	GetHelper(t).Helper()
	values, ok := contextValues.Load(t)
	if !ok {
		switch tt := t.(type) {
		case *tImpl:
			return tt.ctx.Value(key)
		case *inverseTT:
			return tt.parent.Value(key)
		case *eventuallyT:
			return tt.parent.Value(key)
		case *MockT:
			return tt.Parent.Value(key)
		case *testing.T:
			return nil
		default:
			panic(fmt.Sprintf("unrecognized TT type: %T", t))
		}
	}
	return values.(map[any]any)[key]
}

func setValueForT(t T, key, value any) {
	GetHelper(t).Helper()
	values, _ := contextValues.LoadOrStore(t, make(map[any]any))
	if value == nil {
		delete(values.(map[any]any), key)
	} else {
		values.(map[any]any)[key] = value
	}
}
