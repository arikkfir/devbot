package justest

import (
	"context"
	"fmt"
	"testing"
	"time"
)

var (
	ignoredStackTracePrefixes = []string{"testing.", "github.com/arikkfir/devbot/backend/internal/util/testing/justest"}
)

//go:noinline
func NewTT(t T) TT {
	return &tImpl{
		t:   t,
		ctx: context.Background(),
	}
}

type T interface {
	Cleanup(func())
	Fatalf(format string, args ...any)
	Log(args ...any)
	Logf(format string, args ...any)
}

type TT interface {
	T
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any
}

type tImpl struct {
	t   T
	ctx context.Context
}

//go:noinline
func (t *tImpl) Cleanup(f func()) {
	GetHelper(t).Helper()
	t.t.Cleanup(f)
}

//go:noinline
func (t *tImpl) Fatalf(format string, args ...any) {
	GetHelper(t).Helper()
	t.t.Fatalf(format, args...)
}

//go:noinline
func (t *tImpl) Log(args ...any) {
	GetHelper(t).Helper()
	t.t.Log(args...)
}

//go:noinline
func (t *tImpl) Logf(format string, args ...any) {
	GetHelper(t).Helper()
	t.t.Logf(format, args...)
}

//go:noinline
func (t *tImpl) GetHelper() Helper {
	if hp, ok := t.t.(HelperProvider); ok {
		return hp.GetHelper()
	} else if h, ok := t.t.(Helper); ok {
		return h
	} else {
		panic(fmt.Sprintf("could not obtain a HelperProvider instance from: %+v", t.t))
	}
}

//go:noinline
func (t *tImpl) Deadline() (deadline time.Time, ok bool) {
	GetHelper(t).Helper()
	return t.ctx.Deadline()
}

//go:noinline
func (t *tImpl) Done() <-chan struct{} {
	GetHelper(t).Helper()
	return t.ctx.Done()
}

//go:noinline
func (t *tImpl) Err() error {
	GetHelper(t).Helper()
	return t.ctx.Err()
}

//go:noinline
func (t *tImpl) Value(key any) interface{} {
	GetHelper(t).Helper()
	return getValueForT(t, key)
}

//goland:noinspection GoUnusedExportedFunction
func RootOf(t T) *testing.T {
	switch tt := t.(type) {
	case *tImpl:
		return RootOf(tt.t)
	case *inverseTT:
		return RootOf(tt.parent)
	case *eventuallyT:
		return RootOf(tt.parent)
	case *MockT:
		return RootOf(tt.Parent)
	case *testing.T:
		return tt
	default:
		panic(fmt.Sprintf("unrecognized TT type: %T", t))
	}
}
