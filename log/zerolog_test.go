package log_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/sabariramc/go-kit/log"
	"github.com/sabariramc/go-kit/log/correlation"
)

func TestZerolog(t *testing.T) {
	os.Setenv("LOG_LEVEL", "trace")
	jlog := log.New("default")
	ctx := correlation.GetContextWithCorrelationParam(context.Background(), &correlation.EventCorrelation{
		CorrelationID: "12345",
		ScenarioID:    "67890",
		SessionID:     "abcde"})

	jlog.Debug().Ctx(ctx).Msg("This is a debug message")
	clog := log.NewConsoleWriter("test")
	clog.Debug().Ctx(ctx).Msg("This is a debug message in console")
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
			jlog.Debug().Ctx(ctx).Msg("This is a debug message")
		}
	})
	nctx := context.Background()
	jlog = log.New("default", log.WithTarget(io.Discard), log.WithNewHooks()).With().Dict("correlation", zerolog.Dict().
		Str("correlationId", "12345").Str("scenarioId", "67890").Str("sessionId", "abcd")).Logger()
	b.Run("without Context", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jlog.Debug().Ctx(nctx).Msg("This is a debug message")
		}
	})
	clog := log.NewConsoleWriter("test", log.WithTarget(io.Discard))

	b.Run("console with Context", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			clog.Debug().Ctx(ctx).Msg("This is a debug message")
		}
	})
	b.Run("console without Context", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			clog.Debug().Ctx(nctx).Msg("This is a debug message")
		}
	})
}
