package log_test

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/sabariramc/go-kit/log"
	"github.com/sabariramc/go-kit/log/correlation"
)

func TestZerolog(t *testing.T) {
	os.Setenv("LOG_LEVEL", "trace")
	jlog := log.New("default", func(c *log.Config) {
		c.LevelScanner = 100 * time.Millisecond
	})
	ctx := correlation.GetContextWithCorrelationParam(context.Background(), &correlation.EventCorrelation{
		CorrelationID: "12345",
		ScenarioID:    "67890",
		SessionID:     "abcde"})

	jlog.Debug(ctx).Msg("This is a debug message")
	clog := log.NewConsoleWriter("test", func(c *log.Config) {
		c.LevelScanner = 100 * time.Millisecond
	})
	clog.Debug(ctx).Msg("This is a debug message in console")
	os.Setenv("LOG_LEVEL", "error")
	time.Sleep(200 * time.Millisecond) // Wait for level scanner to update the log level
	jlog.Debug(ctx).Msg("This is a debug message")
	jlog.Error(ctx).Msg("This is an error message")
	clog.Debug(ctx).Msg("This is a debug message in console")
	clog.Error(ctx).Msg("This is an error message in console")
	// Output:
	// {"level":"debug","module":"default","time":"2025-07-24T18:01:18+05:30","correlation":{"correlationID":"12345","scenarioID":"67890","sessionID":"abcde"},"message":"This is a debug message"}
	// 2025-07-24T18:01:18+05:30 DBG This is a debug message in console correlation={"correlationID":"12345","scenarioID":"67890","sessionID":"abcde"} module=test
	// {"level":"error","module":"default","time":"2025-07-24T18:01:19+05:30","correlation":{"correlationID":"12345","scenarioID":"67890","sessionID":"abcde"},"message":"This is an error message"}
	// 2025-07-24T18:01:19+05:30 ERR This is an error message in console correlation={"correlationID":"12345","scenarioID":"67890","sessionID":"abcde"} module=test
}

func BenchmarkLog(b *testing.B) {
	os.Setenv("LOG_LEVEL", "trace")
	jlog := log.New("default", log.WithTarget(io.Discard))
	ctx := correlation.GetContextWithCorrelationParam(context.Background(), &correlation.EventCorrelation{
		CorrelationID: "12345",
		ScenarioID:    "67890",
		SessionID:     "abcde"})
	b.Run("with Context", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jlog.Debug(ctx).Msg("This is a debug message")
		}
	})
	nctx := context.Background()
	jlog = &log.Logger{log.New("default", log.WithTarget(io.Discard), log.WithNewHooks()).With().Dict("correlation", zerolog.Dict().
		Str("correlationId", "12345").Str("scenarioId", "67890").Str("sessionId", "abcd")).Logger(),
	}
	b.Run("without Context", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jlog.Debug(nctx).Msg("This is a debug message")
		}
	})
	clog := log.NewConsoleWriter("test", log.WithTarget(io.Discard))

	b.Run("console with Context", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			clog.Debug(ctx).Msg("This is a debug message")
		}
	})
	b.Run("console without Context", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			clog.Debug(nctx).Msg("This is a debug message")
		}
	})
}
