package util

import (
	"fmt"
	"testing"
	"time"
)

type TestingT interface {
	Error(args ...any)
	Errorf(format string, args ...any)
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Log(args ...any)
	Logf(format string, args ...any)
	TempDir() string
}

type tWrapper struct {
	*testing.T
	e []error
}

type fatalSignal func(t *testing.T)

func (w *tWrapper) Error(args ...interface{}) {
	w.T.Helper()
	w.e = append(w.e, fmt.Errorf(fmt.Sprintln(args...)))
}

func (w *tWrapper) Errorf(format string, args ...interface{}) {
	w.T.Helper()
	w.e = append(w.e, fmt.Errorf(format, args...))
}

func (w *tWrapper) Fail() {
	w.T.Helper()
	err := fmt.Errorf("failed")
	w.e = append(w.e, err)
}

func (w *tWrapper) FailNow() {
	w.T.Helper()
	panic(fatalSignal(func(t *testing.T) { t.FailNow() }))
}

func (w *tWrapper) Failed() bool {
	w.T.Helper()
	return len(w.e) > 0
}

func (w *tWrapper) Fatal(args ...interface{}) {
	w.T.Helper()
	panic(fatalSignal(func(t *testing.T) { t.Fatal(args...) }))
}

func (w *tWrapper) Fatalf(format string, args ...interface{}) {
	w.T.Helper()
	panic(fatalSignal(func(t *testing.T) { t.Fatalf(format, args...) }))
}

func (w *tWrapper) Log(args ...interface{}) {
	w.T.Helper()
	w.T.Log(args...)
}

func (w *tWrapper) Logf(format string, args ...interface{}) {
	w.T.Helper()
	w.T.Logf(format, args...)
}

func (w *tWrapper) TempDir() string {
	w.T.Helper()
	return w.T.TempDir()
}

func Eventually(t *testing.T, waitFor time.Duration, interval time.Duration, f func(t TestingT)) {
	t.Helper()

	timer := time.NewTimer(waitFor)
	defer timer.Stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	w := &tWrapper{T: t}

	panicCh := make(chan interface{}, 1)
	fatalCh := make(chan fatalSignal, 1)
	doneCh := make(chan interface{}, 1)
	var fatal fatalSignal
	for tick := ticker.C; ; {
		select {
		case <-timer.C:
			if !w.Failed() {
				w.Errorf("Timed out")
			}
			for _, err := range w.e {
				t.Errorf("%v", err)
			}
			if fatal != nil {
				// evaluation function called t.Fatal* or t.FailNow - apply it now to the upstream "t" which will panic
				fatal(t)
			} else {
				// we're done (whether evaluation function called t.Error* or not)
				return
			}
		case <-tick:
			tick = nil
			w.e = nil
			go func() {
				defer func() {
					if r := recover(); r != nil {
						if p, ok := r.(fatalSignal); ok {
							fatalCh <- p
						} else {
							panicCh <- r
						}
					}
				}()
				f(w)
				doneCh <- nil
			}()
		case r := <-panicCh:
			// given evaluation function panicked unexpectedly (not a t.Fatal* call)
			// in such cases, we bubble the panic upstream since this is not an expected failure
			panic(r)
		case r := <-fatalCh:
			// evaluation function called t.Fatal* or t.FailNow - record it but keep trying
			fatal = r
			tick = ticker.C
		case <-doneCh:
			// evaluation function finished, but may have called t.Error* - if so, we'll continue trying; return otherwise
			fatal = nil
			if w.Failed() {
				tick = ticker.C
			} else {
				return
			}
		}
	}
}
