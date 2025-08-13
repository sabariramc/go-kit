package kafka

import (
	"context"
	"fmt"

	"github.com/sabariramc/go-kit/app/base"
	span "github.com/sabariramc/go-kit/instrumentation"
	"github.com/sabariramc/go-kit/kafka/consumer"
	"github.com/sabariramc/go-kit/log"
	"github.com/segmentio/kafka-go"
)

type Tracer interface {
	span.SpanOp
	StartKafkaSpanFromMessage(ctx context.Context, msg *kafka.Message) (context.Context, span.Span)
}

type Config struct {
	Reader             *consumer.Reader
	Base               *base.Base
	Log                *log.Logger
	Tracer             Tracer
	MessageChannelSize int
}

func NewConfig(opt ...Options) (*Config, error) {
	cfg := &Config{
		Base:               base.New(),
		Log:                log.New("KafkaConsumer"),
		MessageChannelSize: 1,
	}
	for _, o := range opt {
		if err := o(cfg); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}
	if cfg.Reader == nil {
		consumer, err := consumer.New(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
		}
		cfg.Reader = consumer
	}
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return cfg, nil
}

func ValidateConfig(cfg *Config) error {
	if cfg.Reader == nil {
		return fmt.Errorf("kafka reader is not configured")
	}
	if cfg.Base == nil {
		return fmt.Errorf("base is not configured")
	}
	if cfg.Log == nil {
		return fmt.Errorf("logger is not configured")
	}
	return nil
}

type Options func(*Config) error
