package producer

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/sabariramc/go-kit/env"
	span "github.com/sabariramc/go-kit/instrumentation"
	ck "github.com/sabariramc/go-kit/kafka"
	"github.com/sabariramc/go-kit/log"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
)

const (
	ModuleProducer = "KafkaProducer"
)

type Config struct {
	Log    *log.Logger
	Writer *kafka.Writer
	Hooks  []Hook
	Tracer span.SpanOp
}

func ValidateConfig(config *Config) error {
	if config.Writer == nil {
		return fmt.Errorf("kafka writer is not initialized")
	}
	if config.Writer.Async {
		config.Log.Info(context.Background()).Msgf("Kafka writer is set to Async mode")
		if config.Writer.AllowAutoTopicCreation {
			config.Log.Warn(context.Background()).Msgf("Kafka writer is set to Async mode with AllowAutoTopicCreation enabled, this may lead to unexpected topic creation")
		}
	}
	if config.Writer.RequiredAcks == 0 {
		config.Log.Info(context.Background()).Msgf("Kafka replica acknowledgement is set to None")
	}

	return nil
}

func completionReport(log *log.Logger) func(messages []kafka.Message, err error) {
	return func(messages []kafka.Message, err error) {
		if err != nil {
			log.Error(context.TODO()).Err(err).Msg("Failed to write messages to topic")
		}
		if len(messages) != 0 {
			log.Debug(context.TODO()).Dict("topic", zerolog.Dict().Str("name", messages[0].Topic).Str("partition", strconv.Itoa(messages[0].Partition))).Msgf("Successfully wrote %d messages to topic", len(messages))
		}
	}
}

func NewDefaultConfig() *Config {
	logger := log.New(ModuleProducer)
	completionReportLog := log.New(ModuleProducer, func(c *log.Config) {
		c.Labels = map[string]string{"type": "completion_report"}
	})
	internalLog := log.New(ModuleProducer, func(c *log.Config) {
		c.Labels = map[string]string{"type": "internal_log"}
	})
	config := &Config{
		Log:   logger,
		Hooks: []Hook{HookFunc(CorelationHook)},
		Writer: &kafka.Writer{
			Addr:         kafka.TCP(env.GetSlice(ck.EnvBroker, []string{"localhost:9093"}, ",")...),
			Topic:        env.Get(ck.EnvProducerTopic, ""),
			Balancer:     &kafka.Hash{},
			Completion:   completionReport(completionReportLog),
			RequiredAcks: kafka.RequiredAcks(env.GetInt(ck.EnvProducerAcknowledge, 1)),
			Transport: &kafka.Transport{
				Dial: (&net.Dialer{
					Timeout:   3 * time.Second,
					DualStack: true,
				}).DialContext,
			},
			Logger: kafka.LoggerFunc(func(s string, i ...interface{}) {
				internalLog.Debug(context.TODO()).Msgf(s, i...)
			}),
			ErrorLogger: kafka.LoggerFunc(func(s string, i ...interface{}) {
				internalLog.Error(context.TODO()).Msgf(s, i...)
			}),
			Async:                  env.GetBool(ck.EnvProducerAsync, true),
			AllowAutoTopicCreation: false, // Note: if async is true, this should be set to false to avoid unexpected topic creation
		},
	}
	return config
}

type Options func(*Config) error

func WithLogger(logger *log.Logger) Options {
	return func(c *Config) error {
		c.Log = logger
		internalLog := log.New(ModuleProducer, func(c *log.Config) {
			c.Labels = map[string]string{"type": "internal_log"}
			c.Logger = &logger.Logger
		})
		completionReportLog := log.New(ModuleProducer, func(c *log.Config) {
			c.Labels = map[string]string{"type": "completion_report"}
			c.Logger = &logger.Logger
		})
		if c.Writer != nil {
			c.Writer.Completion = completionReport(completionReportLog)
			c.Writer.Logger = kafka.LoggerFunc(func(s string, i ...interface{}) {
				internalLog.Debug(context.TODO()).Msgf(s, i...)
			})
			c.Writer.ErrorLogger = kafka.LoggerFunc(func(s string, i ...interface{}) {
				internalLog.Error(context.TODO()).Msgf(s, i...)
			})
		}
		return nil
	}
}

func WithHooks(hooks ...Hook) Options {
	return func(c *Config) error {
		c.Hooks = append(c.Hooks, hooks...)
		return nil
	}
}

func WithTracer(tracer Tracer) Options {
	return func(c *Config) error {
		c.Hooks = append(c.Hooks, TraceHook{tr: tracer})
		c.Tracer = tracer
		return nil
	}
}

func WithPlainSSLMechanism(username, password string) Options {
	return func(c *Config) error {
		if c.Writer != nil && c.Writer.Transport != nil {
			kafkaTransport, ok := c.Writer.Transport.(*kafka.Transport)
			if !ok {
				return fmt.Errorf("kafka transport is not set")
			}
			kafkaTransport.SASL = &plain.Mechanism{Username: username, Password: password}
		}
		return nil
	}
}

func WithSSLMechanism(mechanism sasl.Mechanism) Options {
	return func(c *Config) error {
		if c.Writer != nil && c.Writer.Transport != nil {
			kafkaTransport, ok := c.Writer.Transport.(*kafka.Transport)
			if !ok {
				return fmt.Errorf("kafka transport is not set")
			}
			kafkaTransport.SASL = mechanism
		}
		return nil
	}
}

func WithoutInternalLogger() Options {
	return func(c *Config) error {
		if c.Writer != nil {
			c.Writer.Logger = nil
			c.Writer.ErrorLogger = nil
		}
		return nil
	}
}
