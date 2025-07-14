package log_test

import (
	"context"
	"os"
	"testing"

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
