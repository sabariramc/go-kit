package log

import (
	"io"
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

func NewConsoleWriter(module string, opt ...Option) zerolog.Logger {
	cfg := NewConfig(opt...)
	return zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.RFC3339
		w.Out = cfg.Target
		w.NoColor = false
	})).With().Timestamp().Str("module", module).Logger().Level(cfg.Level).Hook(cfg.Hooks...)
}

func New(module string, opt ...Option) zerolog.Logger {
	cfg := NewConfig(opt...)
	return zerolog.New(cfg.Target).With().
		Timestamp().Str("module", module).Logger().Level(cfg.Level).Hook(cfg.Hooks...)
}

type Config struct {
	Hooks  []zerolog.Hook
	Target io.Writer
	Level  zerolog.Level
}

func NewConfig(opts ...Option) *Config {
	c := &Config{
		Hooks:  []zerolog.Hook{zerolog.HookFunc(EventCorrelation)},
		Target: os.Stdout,
		Level:  getLevel(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Option func(*Config)

func WithHooks(hooks ...zerolog.Hook) Option {
	return func(c *Config) {
		c.Hooks = append(c.Hooks, hooks...)
	}
}

func WithTarget(target io.Writer) Option {
	return func(c *Config) {
		c.Target = target
	}
}

func WithLevel(level zerolog.Level) Option {
	return func(c *Config) {
		c.Level = level
	}
}

func WithNewHooks(hooks ...zerolog.Hook) Option {
	return func(c *Config) {
		c.Hooks = hooks
	}
}
