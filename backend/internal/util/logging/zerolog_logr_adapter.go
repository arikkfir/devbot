package logging

import (
	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type zeroLogLogrAdapter struct {
	l zerolog.Logger
}

func (a *zeroLogLogrAdapter) Init(_ logr.RuntimeInfo) {
	a.l = log.Logger.With().Logger()
}

func (a *zeroLogLogrAdapter) Enabled(level int) bool {
	return level >= int(a.l.GetLevel())
}

func (a *zeroLogLogrAdapter) Info(_ int, msg string, keysAndValues ...interface{}) {
	e := a.l.Info()
	for i := 0; i < len(keysAndValues); i += 2 {
		k := keysAndValues[i]
		v := keysAndValues[i+1]
		e = e.Interface(k.(string), v)
	}
	e.Msg(msg)
}

func (a *zeroLogLogrAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	e := a.l.Error().Err(err)
	for i := 0; i < len(keysAndValues); i += 2 {
		k := keysAndValues[i]
		v := keysAndValues[i+1]
		e = e.Interface(k.(string), v)
	}
	e.Msg(msg)
}

func (a *zeroLogLogrAdapter) WithValues(keysAndValues ...interface{}) logr.LogSink {
	e := a.l.With()
	for i := 0; i < len(keysAndValues); i += 2 {
		k := keysAndValues[i]
		v := keysAndValues[i+1]
		e = e.Interface(k.(string), v)
	}
	return &zeroLogLogrAdapter{l: e.Logger()}
}

func (a *zeroLogLogrAdapter) WithName(_ string) logr.LogSink {
	return a
}
