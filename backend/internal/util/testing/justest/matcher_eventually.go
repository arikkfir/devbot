package justest

import (
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

//go:noinline
func (b *eventuallyBuilder) Within(timeout time.Duration) EventuallyBuilderWithTimeout {
	if timeout == 0 {
		panic("timeout cannot be 0")
	}
	b.timeout = &timeout
	return b
}

//go:noinline
func (b *eventuallyBuilder) ProbingEvery(interval time.Duration) Matcher {
	if interval == 0 {
		panic("interval cannot be 0")
	} else if b.timeout.Nanoseconds() < interval.Nanoseconds() {
		panic("interval cannot be greater than timeout")
	}
	b.interval = &interval

	tick := func(t *eventuallyT, actuals ...any) []any {
		GetHelper(t).Helper()

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

		defer func() {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Trial %d failed during cleanups: %+v", t.trial, r)
				}
			}()

			for i := len(t.cleanup) - 1; i >= 0; i-- {
				t.cleanup[i]()
			}
		}()

		var resolvedActuals []any
		for _, actual := range actuals {
			value, found := eventuallyValueExtractor.ExtractValue(t, actual)
			if found {
				resolvedActuals = append(resolvedActuals, value)
			}
		}
		return b.expectation(t, resolvedActuals...)
	}

	return func(t TT, actuals ...any) []any {
		GetHelper(t).Helper()

		start := time.Now()
		timeout := time.NewTimer(*b.timeout)
		defer timeout.Stop()

		ticker := time.NewTicker(*b.interval)
		defer ticker.Stop()

		// TODO: spawn ticks in goroutines; cancel context & return on when timeout occurs (hopefully goroutines will also exit)

		trial := 0
		var et *eventuallyT
		for {
			select {
			case <-timeout.C:
				verifyNotInterrupted(t)
				if et == nil {
					t.Fatalf("Eventually matcher timed out before any evaluations could take place")
					panic("unreachable")
				} else if et.error == nil {
					panic("Eventually reached an illegal state where it timed out with failed matchers, yet no error was found. This is a bug and should be reported.")
				} else {
					format := "Timed out after %s waiting for expectation to be met: " + et.error.msg
					args := append([]any{time.Since(start).String()}, et.error.args...)
					t.Fatalf(format, args...)
				}
			case <-ticker.C:
				verifyNotInterrupted(t)
				trial = +1
				et = &eventuallyT{parent: t, trial: trial}
				updatedActuals := tick(et, actuals...)
				if et.error == nil {
					return updatedActuals
				}
			}
		}
	}
}

//go:noinline
func Eventually(expectation Matcher) EventuallyBuilder {
	if expectation == nil {
		panic("expectation cannot be nil")
	}
	return &eventuallyBuilder{expectation: expectation}
}
