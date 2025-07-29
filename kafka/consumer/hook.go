package consumer

import (
	"context"
	"strings"

	"github.com/google/uuid"
	span "github.com/sabariramc/go-kit/instrumentation"
	"github.com/sabariramc/go-kit/log/correlation"
	"github.com/segmentio/kafka-go"
)

type MessageContext interface {
	Get(ctx context.Context, msg *kafka.Message) context.Context
}

type MessageContextFunc func(ctx context.Context, msg *kafka.Message) context.Context

func (h MessageContextFunc) Get(ctx context.Context, msg *kafka.Message) context.Context {
	return h(ctx, msg)
}

type Tracer interface {
	span.SpanOp
	ExtractTraceContext(ctx context.Context, msg *kafka.Message) context.Context
}

type TraceMessageContext struct {
	tr Tracer
}

func (h TraceMessageContext) Get(ctx context.Context, msg *kafka.Message) context.Context {
	if h.tr == nil {
		return nil
	}
	return h.tr.ExtractTraceContext(ctx, msg)
}

func CorelationHook(ctx context.Context, msg *kafka.Message) context.Context {
	eventCorrelation := correlation.EventCorrelation{
		CorrelationID: uuid.NewString(),
	}
	for _, header := range msg.Headers {
		key := strings.ToLower(header.Key)
		switch key {
		case "x-correlation-id":
			eventCorrelation.CorrelationID = string(header.Value)
		case "x-scenario-id":
			eventCorrelation.ScenarioID = string(header.Value)
		case "x-session-id":
			eventCorrelation.SessionID = string(header.Value)
		case "x-scenario-name":
			eventCorrelation.ScenarioName = string(header.Value)
		}
	}
	ctx = correlation.GetContextWithCorrelationParam(ctx, &eventCorrelation)
	return ctx
}
