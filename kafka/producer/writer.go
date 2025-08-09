package producer

import (
	"context"
	"fmt"

	span "github.com/sabariramc/go-kit/instrumentation"
	"github.com/sabariramc/go-kit/log"
	"github.com/segmentio/kafka-go"
)

type Writer struct {
	*kafka.Writer
	log   *log.Logger
	hooks []Hook
	tr    span.SpanOp
}

func New(ctx context.Context, opt ...Options) (*Writer, error) {
	cfg := NewDefaultConfig()
	for _, o := range opt {
		err := o(cfg)
		if err != nil {
			return nil, fmt.Errorf("kafka.Writer.New: error applying option: %w", err)
		}
	}
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("kafka.Writer.New: validation error: %w", err)
	}
	return &Writer{
		Writer: cfg.Writer,
		log:    cfg.Log,
		hooks:  cfg.Hooks,
		tr:     cfg.Tracer,
	}, nil
}

func (w *Writer) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if len(msgs) == 0 {
		return nil
	}
	for i := range msgs {
		w.process(ctx, &msgs[i])
	}
	if w.tr != nil {
		var sp span.Span
		ctx, sp = w.tr.NewSpanFromContext(ctx, "kafka.write", span.SpanKindProducer, "")
		defer sp.Finish()
	}
	return w.Writer.WriteMessages(ctx, msgs...)
}

func (w *Writer) process(ctx context.Context, msg *kafka.Message) {
	for _, hook := range w.hooks {
		if err := hook.Run(ctx, msg); err != nil {
			w.log.Error(ctx).Err(err).Msgf("Failed to execute hook for message with key %s", string(msg.Key))
		}
	}
}

func (w *Writer) Close(ctx context.Context) error {
	if err := w.Writer.Close(); err != nil {
		return fmt.Errorf("kafka.Writer.Close: error closing writer: %w", err)
	}
	return nil
}
