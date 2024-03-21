package justest

import (
	"context"
	"fmt"
	"github.com/secureworks/errors"
	"reflect"
	"time"
)

var (
	eventuallyValueExtractor ValueExtractor
)

func init() {
	eventuallyValueExtractor = NewValueExtractor(ExtractSameValue)
	eventuallyValueExtractor[reflect.Chan] = NewChannelExtractor(eventuallyValueExtractor, false)
	eventuallyValueExtractor[reflect.Func] = NewFuncExtractor(eventuallyValueExtractor, false)
	eventuallyValueExtractor[reflect.Pointer] = NewPointerExtractor(eventuallyValueExtractor, false)
}

type eventuallyMatcher struct {
	expectation Matcher
	timeout     time.Duration
	interval    time.Duration
}

func (m *eventuallyMatcher) ExpectMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()

	timeout := time.NewTimer(m.timeout)
	defer timeout.Stop()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	expired := false
	for {
		select {
		case <-timeout.C:
			verifyNotInterrupted(t)
			expired = true
		case <-ticker.C:
			verifyNotInterrupted(t)
			tt := &eventuallyJustT{parent: t, expired: expired}
			updatedActuals := m.tick(context.Background(), tt, actuals...)
			if tt.error == nil {
				return updatedActuals
			}
		}
	}
}

func (m *eventuallyMatcher) ExpectNoMatch(t JustT, actuals ...any) []any {
	GetHelper(t).Helper()

	timeout := time.NewTimer(m.timeout)
	defer timeout.Stop()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			verifyNotInterrupted(t)
			return actuals
		case <-ticker.C:
			verifyNotInterrupted(t)
			tt := &eventuallyJustT{parent: t, expired: false}
			_ = m.tick(context.Background(), tt, actuals...)
			if tt.error == nil {
				t.Fatalf("Expected condition to not be met, but it was eventually met")
				panic("unreachable")
			}
		}
	}
}

func (m *eventuallyMatcher) tick(ctx context.Context, t *eventuallyJustT, actuals ...any) []any {
	GetHelper(t).Helper()
	if !t.expired {
		defer func() {
			if r := recover(); r != nil {
				if r != t {
					if err, isErr := r.(error); isErr {
						panic(errors.New("unexpected panic! recovered: %w", err))
					} else {
						panic(errors.New("unexpected panic! recovered: %+v", r))
					}
				}
			}
		}()
	}

	defer func() {
		for i := len(t.cleanup) - 1; i >= 0; i-- {
			t.cleanup[i]()
		}
	}()

	resolvedActuals := eventuallyValueExtractor.ExtractValues(ctx, t, actuals...)
	return m.expectation.ExpectMatch(t, resolvedActuals...)
}

type EventuallyBuilder interface {
	Within(timeout time.Duration) EventuallyBuilderWithTimeout
}

type EventuallyBuilderWithTimeout interface {
	ProbingEvery(interval time.Duration) Matcher
}

type eventuallyBuilder struct {
	expectation Matcher
	timeout     *time.Duration
	interval    *time.Duration
}

func (b *eventuallyBuilder) Within(timeout time.Duration) EventuallyBuilderWithTimeout {
	if timeout == 0 {
		panic("timeout cannot be 0")
	}
	b.timeout = &timeout
	return b
}

func (b *eventuallyBuilder) ProbingEvery(interval time.Duration) Matcher {
	if interval == 0 {
		panic("interval cannot be 0")
	} else if b.timeout.Nanoseconds() < interval.Nanoseconds() {
		panic("interval cannot be greater than timeout")
	}
	b.interval = &interval
	return &eventuallyMatcher{expectation: b.expectation, timeout: *b.timeout, interval: *b.interval}
}

func Eventually(expectation Matcher) EventuallyBuilder {
	if expectation == nil {
		panic("expectation cannot be nil")
	}
	return &eventuallyBuilder{expectation: expectation}
}

type eventuallyJustT struct {
	parent  JustT
	expired bool
	cleanup []func()
	error   *struct {
		msg  string
		args []any
	}
}

func (t *eventuallyJustT) Cleanup(f func()) {
	GetHelper(t).Helper()
	t.cleanup = append(t.cleanup, f)
}

func (t *eventuallyJustT) GetHelper() Helper {
	if hp, ok := t.parent.(HelperProvider); ok {
		return hp.GetHelper()
	} else if h, ok := t.parent.(Helper); ok {
		return h
	} else {
		panic(fmt.Sprintf("could not obtain a HelperProvider instance from the given JustT instance: %+v", t.parent))
	}
}

func (t *eventuallyJustT) Log(args ...any) {
	GetHelper(t).Helper()
	t.parent.Log(args...)
}

func (t *eventuallyJustT) Logf(format string, args ...any) {
	GetHelper(t).Helper()
	t.parent.Logf(format, args...)
}

func (t *eventuallyJustT) Fatalf(format string, args ...any) {
	GetHelper(t).Helper()
	if t.expired {
		t.parent.Fatalf(format, args...)
	} else {
		t.error = &struct {
			msg  string
			args []any
		}{msg: format, args: args}
		panic(t)
	}
}
