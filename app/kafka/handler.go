package kafka

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/sabariramc/go-kit/app/base"
	span "github.com/sabariramc/go-kit/instrumentation"
	"github.com/sabariramc/go-kit/log/correlation"
	"github.com/segmentio/kafka-go"
)

func (k *KafkaConsumer) AddHandler(ctx context.Context, topicName string, handler Handler) {
	if handler == nil {
		k.log.Panic(ctx).Err(fmt.Errorf("KafkaConsumer.AddHandler: handler parameter cannot be nil")).Msg("missing handler for topic - " + topicName)
	}
	if _, ok := k.handler[topicName]; ok {
		k.log.Panic(ctx).Err(fmt.Errorf("KafkaConsumer.AddHandler: handler for topic exist")).Msg("duplicate handler for topic - " + topicName)
	}
	if _, ok := k.topics[topicName]; !ok {
		k.log.Panic(ctx).Err(fmt.Errorf("KafkaConsumer.AddHandler: topic not subscribed")).Msg("topic not subscribed - " + topicName)
	}
	k.handler[topicName] = handler
}

func (k *KafkaConsumer) Handle(ctx context.Context, msg *kafka.Message) {
	var span span.Span
	var statusCode = http.StatusOK
	var err error
	topicName := msg.Topic
	handler := k.handler[topicName]
	if handler == nil {
		k.log.Error(ctx).Msg("missing handler for topic - " + topicName)
	}
	if k.tr != nil {
		span, ctx = k.startSpan(ctx, msg)
		if span != nil {
			defer func() {
				if err != nil {
					span.SetError(err, "")
					statusCode, _ = base.ProcessError(ctx, err)
				}
				span.SetStatus(statusCode, http.StatusText(statusCode))
				span.Finish()
			}()
		}
	}
	defer func() {
		if rec := recover(); rec != nil {
			stackTrace := string(debug.Stack())
			k.log.Error(ctx).Any("panic", rec).Str("stacktrace", stackTrace).Msg("Panic recovered")
			var ok bool
			err, ok = rec.(error)
			if !ok {
				err = fmt.Errorf("error occurred during request processing")
			}
		}
	}()
	err = handler.Handle(ctx, msg)
	if err != nil {
		k.log.Error(ctx).Err(err).Msg("Error in processing kafka message")
	}
}

func (k *KafkaConsumer) startSpan(ctx context.Context, msg *kafka.Message) (span.Span, context.Context) {
	if k.tr != nil {
		corr, _ := correlation.ExtractCorrelationParam(ctx)
		msgCtx, span := k.tr.StartKafkaSpanFromMessage(ctx, msg)
		span.SetAttribute("correlationId", corr.CorrelationID)
		span.SetAttribute("messaging.kafka.topic", msg.Topic)
		span.SetAttribute("messaging.kafka.partition", msg.Partition)
		span.SetAttribute("messaging.kafka.offset", msg.Offset)
		span.SetAttribute("messaging.kafka.key", string(msg.Key))
		span.SetAttribute("messaging.kafka.timestamp", msg.Time.UnixMilli())
		return span, msgCtx
	}
	return nil, ctx
}
