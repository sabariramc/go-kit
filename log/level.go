package log

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/sabariramc/go-kit/env"
)

func getLevel() zerolog.Level {
	logLevel := env.Get("LOG_LEVEL", "error")
	lvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		lvl = zerolog.ErrorLevel
	}
	return lvl
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
