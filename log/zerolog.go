package log

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog.Logger
}

func New(module string, opt ...Option) *Logger {
	cfg := NewConfig(opt...)
	if cfg.Logger == nil {
		lg := zerolog.New(cfg.Target)
		cfg.Logger = &lg
	}
	logCtx := cfg.Logger.With()
	for key, value := range cfg.Labels {
		logCtx = logCtx.Str(key, value)
	}
	logCtx = logCtx.Str("module", module).Timestamp()
	l := Logger{logCtx.Logger().Level(cfg.Level).Hook(cfg.Hooks...)}
	if cfg.LevelScanner > 0 {
		go l.scanLevel(cfg.LevelScanner)
	}
	return &l
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

func (l *Logger) scanLevel(timeout time.Duration) {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()
	for range ticker.C {
		newLevel := getLevel()
		if newLevel != l.GetLevel() {
			l.Logger = l.Logger.Level(newLevel)
			l.Info(context.Background()).Msgf("Log level changed to %s", newLevel)
		}
	}
}
