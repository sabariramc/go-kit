package log

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog.Logger
}

func (l *Logger) Trace(ctx context.Context) *zerolog.Event {
	return l.Logger.Trace().Ctx(ctx)
}

func (l *Logger) Debug(ctx context.Context) *zerolog.Event {
	return l.Logger.Debug().Ctx(ctx)
}

func (l *Logger) Info(ctx context.Context) *zerolog.Event {
	return l.Logger.Info().Ctx(ctx)
}

func (l *Logger) Warn(ctx context.Context) *zerolog.Event {
	return l.Logger.Warn().Ctx(ctx)
}

func (l *Logger) Error(ctx context.Context) *zerolog.Event {
	return l.Logger.Error().Ctx(ctx)
}

func (l *Logger) Panic(ctx context.Context) *zerolog.Event {
	return l.Logger.Panic().Ctx(ctx)
}

func (l *Logger) Fatal(ctx context.Context) *zerolog.Event {
	return l.Logger.Fatal().Ctx(ctx)
}

func setContext(logCtx zerolog.Context, module string, labels map[string]string) zerolog.Context {
	for key, value := range labels {
		logCtx = logCtx.Str(key, value)
	}
	logCtx = logCtx.Str("module", module).Timestamp()
	return logCtx
}

func NewConsoleWriter(module string, opt ...Option) *Logger {
	cfg := NewConfig(opt...)
	logCtx := zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.RFC3339
		w.Out = cfg.Target
		w.NoColor = false
	})).With()
	logCtx = setContext(logCtx, module, cfg.Labels)
	l := Logger{logCtx.Logger().Level(cfg.Level).Hook(cfg.Hooks...)}
	if cfg.LevelScanner > 0 {
		go l.scanLevel(cfg.LevelScanner)
	}
	return &l
}

func New(module string, opt ...Option) *Logger {
	cfg := NewConfig(opt...)
	logCtx := zerolog.New(cfg.Target).With().Str("module", module).Timestamp()
	logCtx = setContext(logCtx, module, cfg.Labels)
	l := Logger{logCtx.Logger().Level(cfg.Level).Hook(cfg.Hooks...)}
	if cfg.LevelScanner > 0 {
		go l.scanLevel(cfg.LevelScanner)
	}
	return &l
}
