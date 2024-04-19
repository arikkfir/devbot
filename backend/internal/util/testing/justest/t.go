package justest

import "time"

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

type Cleaner interface {
	PerformCleanups()
}
