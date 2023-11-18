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

type panicSignal func(t *testing.T)

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
	panic(panicSignal(func(t *testing.T) { t.FailNow() }))
}

func (w *tWrapper) Failed() bool {
	w.T.Helper()
	return len(w.e) > 0
}

func (w *tWrapper) Fatal(args ...interface{}) {
	w.T.Helper()
	panic(panicSignal(func(t *testing.T) { t.Fatal(args...) }))
}

func (w *tWrapper) Fatalf(format string, args ...interface{}) {
	w.T.Helper()
	panic(panicSignal(func(t *testing.T) { t.Fatalf(format, args...) }))
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

	ch := make(chan interface{}, 1)
	panicked := false
	for tick := ticker.C; ; {
		select {
		case <-timer.C:
			if !w.Failed() {
				w.Errorf("Timed out")
			}
			for _, err := range w.e {
				t.Errorf("%v", err)
			}
			if panicked {
				t.FailNow()
			} else {
				return
			}
		case <-tick:
			tick = nil
			w.e = nil
			go func() {
				defer func() {
					if r := recover(); r != nil {
						ps, ok := r.(panicSignal)
						if ok {
							ch <- ps
							return
						}

						//goland:noinspection GoTypeAssertionOnErrors
						err, ok := r.(error)
						if !ok {
							err = fmt.Errorf("%v", r)
						}
						w.e = append(w.e, err)

						panicked = true
						tick = ticker.C
					}
				}()
				f(w)
				ch <- true
			}()
		case r := <-ch:
			if ps, ok := r.(panicSignal); ok {
				ps(t)
				return
			} else if w.Failed() {
				panicked = false
				tick = ticker.C
			} else {
				return
			}
		}
	}
}
