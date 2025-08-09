package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sabariramc/go-kit/env"
	span "github.com/sabariramc/go-kit/instrumentation"
	ck "github.com/sabariramc/go-kit/kafka"
	"github.com/sabariramc/go-kit/log"
	"github.com/sabariramc/go-kit/log/correlation"
	"github.com/sabariramc/go-kit/log/ratelimiter"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
)

const (
	ModuleConsumer = "KafkaConsumer"
)

type Config struct {
	*kafka.ReaderConfig
	AutoCommit  AutoCommit
	Log         *log.Logger
	Hooks       []Hook
	SpanOp      span.SpanOp
	ClosePollCh bool
}

func ValidateConfig(config *Config) error {
	if config.ReaderConfig == nil {
		return fmt.Errorf("ReaderConfig is required")
	}
	if config.AutoCommit.Enabled && config.AutoCommit.IntervalInMs == 0 {
		return fmt.Errorf("AutoCommit.IntervalInMs must be set when AutoCommit is enabled")
	}
	return config.ReaderConfig.Validate()
}

func NewConfig() *Config {
	rl, _ := ratelimiter.New(func(c *ratelimiter.Config) {
		c.BlockSize = 1 * time.Second
		c.WindowSize = time.Minute
	})
	logger := log.New(ModuleConsumer)
	internalLog := log.New(ModuleConsumer, func(c *log.Config) {
		c.Labels = map[string]string{"type": "internal_log"}
	}, log.WithHooks(rl))
	groupID := env.Get(ck.EnvConsumerGroupID, "cg-kafka-consumer")
	clientID := groupID + "-" + uuid.NewString()
	ctx := correlation.GetContextWithCorrelationParam(context.TODO(), &correlation.EventCorrelation{
		CorrelationID: clientID,
	})
	config := &Config{
		ReaderConfig: &kafka.ReaderConfig{
			Brokers:           env.GetSlice(ck.EnvBroker, []string{"localhost:9093"}, ","),
			GroupID:           groupID,
			GroupTopics:       env.GetSlice(ck.EnvConsumerTopics, []string{}, ","),
			HeartbeatInterval: time.Second,
			QueueCapacity:     int(env.GetInt(ck.EnvConsumerMaxBuffer, 100)),
			Dialer: &kafka.Dialer{
				Timeout:   10 * time.Second,
				ClientID:  clientID,
				DualStack: true,
			},
			Logger: kafka.LoggerFunc(func(s string, i ...interface{}) {
				internalLog.Debug(ctx).Msgf(s, i...)
			}),
			ErrorLogger: kafka.LoggerFunc(func(s string, i ...interface{}) {
				internalLog.Error(ctx).Msgf(s, i...)
			}),
		},
		AutoCommit: AutoCommit{
			Enabled:      env.GetBool(ck.EnvConsumerAutoCommit, true),
			IntervalInMs: uint64(env.GetInt(ck.EnvConsumerAutoCommitIntervalInMs, 1000)),
			BatchSize:    uint64(env.GetInt(ck.EnvConsumerAutoCommitBatchSize, 50)),
		},
		Log:         logger,
		Hooks:       []Hook{HookFunc(CorelationHook)},
		ClosePollCh: true,
	}
	return config
}

type Options func(*Config) error

func WithLogger(logger *log.Logger) Options {
	return func(c *Config) error {
		c.Log = logger
		internalLog := logger.With().Str("type", "internal_log").Logger()
		if c.ReaderConfig != nil {
			c.ReaderConfig.Logger = kafka.LoggerFunc(func(s string, i ...interface{}) {
				internalLog.Debug().Msgf(s, i...)
			})
			c.ReaderConfig.ErrorLogger = kafka.LoggerFunc(func(s string, i ...interface{}) {
				internalLog.Error().Msgf(s, i...)
			})
		}
		return nil
	}
}

func WithPlainSSLMechanism(username, password string) Options {
	return func(c *Config) error {
		if c.ReaderConfig != nil && c.ReaderConfig.Dialer != nil {
			c.ReaderConfig.Dialer.SASLMechanism = &plain.Mechanism{Username: username, Password: password}
		}
		return nil
	}
}

func WithSSLMechanism(mechanism sasl.Mechanism) Options {
	return func(c *Config) error {
		if c.ReaderConfig != nil && c.ReaderConfig.Dialer != nil {
			c.ReaderConfig.Dialer.SASLMechanism = mechanism
		}
		return nil
	}
}

func WithClientID(clientID string) Options {
	return func(c *Config) error {
		if c.ReaderConfig != nil && c.ReaderConfig.Dialer != nil {
			c.ReaderConfig.Dialer.ClientID = clientID
			internalLog := log.New(ModuleConsumer, func(cfg *log.Config) {
				cfg.Labels = map[string]string{"type": "internal_log"}
				cfg.Logger = &c.Log.Logger
			})
			ctx := correlation.GetContextWithCorrelationParam(context.TODO(), &correlation.EventCorrelation{
				CorrelationID: clientID,
			})
			c.ReaderConfig.Logger = kafka.LoggerFunc(func(s string, i ...interface{}) {
				internalLog.Debug(ctx).Msgf(s, i...)
			})
			c.ReaderConfig.ErrorLogger = kafka.LoggerFunc(func(s string, i ...interface{}) {
				internalLog.Error(ctx).Msgf(s, i...)
			})
		}
		return nil
	}
}

func WithGroupID(groupID string) Options {
	return func(c *Config) error {
		if c.ReaderConfig != nil {
			c.ReaderConfig.GroupID = groupID
			WithClientID(groupID + "-" + uuid.NewString())(c)
		}
		return nil
	}
}

func WithoutInternalLogger() Options {
	return func(c *Config) error {
		if c.ReaderConfig != nil {
			c.ReaderConfig.Logger = nil
			c.ReaderConfig.ErrorLogger = nil
		}
		return nil
	}
}
