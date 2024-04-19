package justest

import (
	"context"
	"fmt"
	"time"
)

//go:noinline
func NewTT(t T) TT {
	tt := &tImpl{
		t:   t,
		ctx: context.Background(),
	}
	t.Cleanup(tt.PerformCleanups)
	return tt
}

type tImpl struct {
	t       T
	ctx     context.Context
	cleanup []func()
}

//go:noinline
func (t *tImpl) PerformCleanups() {
	GetHelper(t).Helper()
	cleanups := t.cleanup
	t.cleanup = nil
	for i := len(cleanups) - 1; i >= 0; i-- {
		cleanups[i]()
	}
}

//go:noinline
func (t *tImpl) Cleanup(f func()) { GetHelper(t).Helper(); t.cleanup = append(t.cleanup, f) }

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
