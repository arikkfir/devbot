package observability

import (
	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
)

type ZeroLogLogrAdapter struct {
	l zerolog.Logger
}

func (a *ZeroLogLogrAdapter) Init(_ logr.RuntimeInfo) {}

func (a *ZeroLogLogrAdapter) Enabled(_ int) bool {
	// Defer decision on whether to log or not to zerolog
	return true
}

func (a *ZeroLogLogrAdapter) Info(level int, msg string, keysAndValues ...interface{}) {
	var e *zerolog.Event
	if level == 0 {
		e = a.l.Info()
	} else if level == 1 {
		e = a.l.Debug()
	} else {
		e = a.l.Trace()
	}
	a.writeEvent(e, msg, keysAndValues...)
}

func (a *ZeroLogLogrAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	a.writeEvent(a.l.Error().Err(err), msg, keysAndValues...)
}

func (a *ZeroLogLogrAdapter) WithValues(keysAndValues ...interface{}) logr.LogSink {
	e := a.l.With()
	for i := 0; i < len(keysAndValues); i += 2 {
		k := keysAndValues[i]
		v := keysAndValues[i+1]
		e = e.Interface(k.(string), v)
	}
	return &ZeroLogLogrAdapter{l: e.Logger()}
}

func (a *ZeroLogLogrAdapter) WithName(_ string) logr.LogSink {
	return a
}

func (a *ZeroLogLogrAdapter) writeEvent(e *zerolog.Event, msg string, keysAndValues ...interface{}) {
	for i := 0; i < len(keysAndValues); i += 2 {
		k := keysAndValues[i]
		v := keysAndValues[i+1]
		e = e.Interface(k.(string), v)
	}
	e.Msg(msg)
}
