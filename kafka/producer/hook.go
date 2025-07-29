package producer

import (
	"context"

	span "github.com/sabariramc/go-kit/instrumentation"
	"github.com/sabariramc/go-kit/log/correlation"
	"github.com/segmentio/kafka-go"
)

type Hook interface {
	Run(ctx context.Context, msg *kafka.Message) error
}

type HookFunc func(ctx context.Context, msg *kafka.Message) error

func (h HookFunc) Run(ctx context.Context, msg *kafka.Message) error {
	return h(ctx, msg)
}

type Tracer interface {
	span.SpanOp
	InjectKafkaTrace(ctx context.Context, msg *kafka.Message)
}

type TraceHook struct {
	tr Tracer
}

func (h TraceHook) Run(ctx context.Context, msg *kafka.Message) error {
	if h.tr == nil {
		return nil
	}
	var crSpan span.Span
	ctx, crSpan = h.tr.NewSpanFromContext(ctx, "kafka.produce", span.SpanKindProducer, msg.Topic)
	crSpan.SetAttribute("messaging.kafka.topic", msg.Topic)
	crSpan.SetAttribute("messaging.kafka.key", string(msg.Key))
	crSpan.SetAttribute("messaging.kafka.timestamp", msg.Time)
	defer crSpan.Finish()
	h.tr.InjectKafkaTrace(ctx, msg)
	return nil
}

func CorelationHook(ctx context.Context, msg *kafka.Message) error {
	correlationParam, ok := correlation.ExtractCorrelationParam(ctx)
	if ok {
		headers := correlationParam.GetHeader()
		for i, v := range headers {
			msg.Headers = append(msg.Headers, kafka.Header{
				Key:   i,
				Value: []byte(v),
			})
		}
	}
	return nil
}
