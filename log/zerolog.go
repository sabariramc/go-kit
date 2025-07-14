package log

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/sabariramc/go-kit/env"
	"github.com/sabariramc/go-kit/log/correlation"
)

func EventCorrelation(e *zerolog.Event, level zerolog.Level, message string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}
	if corr, ok := correlation.ExtractCorrelationParam(ctx); ok && corr != nil {
		msgDict := zerolog.Dict().Str("correlationID", corr.CorrelationID)
		if corr.ScenarioID != "" {
			msgDict = msgDict.Str("scenarioID", corr.ScenarioID)
		}
		if corr.SessionID != "" {
			msgDict = msgDict.Str("sessionID", corr.SessionID)
		}
		if corr.ScenarioName != "" {
			msgDict = msgDict.Str("scenarioName", corr.ScenarioName)
		}
		e.Dict("correlation", msgDict)
	}
}

func getLevel() zerolog.Level {
	logLevel := env.Get("LOG_LEVEL", "error")
	lvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		lvl = zerolog.ErrorLevel
	}
	return lvl
}

func NewConsoleWriter(module string, hooks ...zerolog.Hook) zerolog.Logger {
	return zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.RFC3339
		w.Out = os.Stdout
		w.NoColor = false
	})).With().Timestamp().Str("module", module).Logger().Level(getLevel()).Hook(zerolog.HookFunc(EventCorrelation)).Hook(hooks...)
}

func New(module string, hooks ...zerolog.Hook) zerolog.Logger {
	return zerolog.New(os.Stdout).With().
		Timestamp().Str("module", module).Logger().Level(getLevel()).Hook(zerolog.HookFunc(EventCorrelation)).Hook(hooks...)
}
