package logging

import (
	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ZeroLogLogrAdapter struct {
	l zerolog.Logger
}

func (a *ZeroLogLogrAdapter) Init(_ logr.RuntimeInfo) {
	a.l = log.Logger.With().Logger()
}

func (a *ZeroLogLogrAdapter) Enabled(level int) bool {
	return level >= int(a.l.GetLevel())
}

func (a *ZeroLogLogrAdapter) Info(_ int, msg string, keysAndValues ...interface{}) {
	e := a.l.Info()
	for i := 0; i < len(keysAndValues); i += 2 {
		k := keysAndValues[i]
		v := keysAndValues[i+1]
		e = e.Interface(k.(string), v)
	}
	e.Msg(msg)
}

func (a *ZeroLogLogrAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	e := a.l.Error().Err(err)
	for i := 0; i < len(keysAndValues); i += 2 {
		k := keysAndValues[i]
		v := keysAndValues[i+1]
		e = e.Interface(k.(string), v)
	}
	e.Msg(msg)
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
