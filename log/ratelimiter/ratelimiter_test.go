package ratelimiter_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/sabariramc/go-kit/log"
	"github.com/sabariramc/go-kit/log/correlation"
	"github.com/sabariramc/go-kit/log/ratelimiter"
	"gotest.tools/v3/assert"
)

func TestRateLimittedLogger(t *testing.T) {
	os.Setenv(log.EnvLogLevel, "trace")
	rl, err := ratelimiter.New(func(c *ratelimiter.Config) {
		c.BlockSize = 100 * time.Millisecond
		c.WindowSize = 500 * time.Millisecond
	})
	assert.NilError(t, err)
	logger := log.New("ratelimitter", log.WithConsole(), log.WithHooks(rl))
	ctx := context.Background()
	logger.Info(ctx).Msg("This is a test message")
	logger.Info(ctx).Msg("This is a test message")
	logger.Info(ctx).Msg("This is a test message1")
	logger.Info(ctx).Msg("This is a test message1")
	logger.Info(ctx).Msg("This is a test message2")
	logger.Info(ctx).Msg("This is a test message2")
	logger.Info(ctx).Msg("This is a test message3")
	logger.Info(ctx).Msg("This is a test message3")
	time.Sleep(1 * time.Second) // Wait for the rate limiter to process
	logger.Info(ctx).Msg("This is a test message")
	logger.Info(ctx).Msg("This is a test message")
	logger.Info(ctx).Msg("This is a test message1")
	logger.Info(ctx).Msg("This is a test message1")
	logger.Info(ctx).Msg("This is a test message2")
	logger.Info(ctx).Msg("This is a test message2")
	logger.Info(ctx).Msg("This is a test message3")
	logger.Info(ctx).Msg("This is a test message3")
}

func BenchmarkRateLimiter(b *testing.B) {
	os.Setenv(log.EnvLogLevel, "trace")
	logMessageList := func() []string {
		res := make([]string, 0, 1000)
		for i := 0; i < 1000; i++ {
			res = append(res, fmt.Sprintf("This is a test message%d", i))
		}
		return res
	}
	nextFunc := func() func() string {
		messages := logMessageList()
		i := 0
		return func() string {
			if i >= len(messages) {
				i = 0
			}
			msg := messages[i]
			i++
			return msg
		}
	}
	rl, err := ratelimiter.New(func(c *ratelimiter.Config) {
		c.BlockSize = 10 * time.Millisecond
		c.WindowSize = 50 * time.Millisecond
	})
	assert.NilError(b, err)
	log := log.New("default", log.WithTarget(io.Discard), log.WithHooks(rl))
	b.ResetTimer()
	ctx := correlation.GetContextWithCorrelationParam(context.Background(), &correlation.EventCorrelation{
		CorrelationID: "12345",
		ScenarioID:    "67890",
		SessionID:     "abcde"})
	next := nextFunc()
	for i := 0; i < b.N; i++ {
		log.Info(ctx).Msg(next())
	}
}
