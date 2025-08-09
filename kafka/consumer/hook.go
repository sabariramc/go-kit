package consumer

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/sabariramc/go-kit/log/correlation"
	"github.com/segmentio/kafka-go"
)

type Hook interface {
	Run(ctx context.Context, msg *kafka.Message) context.Context
}

type HookFunc func(ctx context.Context, msg *kafka.Message) context.Context

func (h HookFunc) Run(ctx context.Context, msg *kafka.Message) context.Context {
	return h(ctx, msg)
}

func CorelationHook(ctx context.Context, msg *kafka.Message) context.Context {
	eventCorrelation := correlation.EventCorrelation{
		CorrelationID: uuid.NewString(),
	}
	for _, header := range msg.Headers {
		key := strings.ToLower(header.Key)
		switch strings.ToLower(key) {
		case strings.ToLower(correlation.CorrelationIDHeader):
			eventCorrelation.CorrelationID = string(header.Value)
		case strings.ToLower(correlation.CorrelationIDHeader):
			eventCorrelation.ScenarioID = string(header.Value)
		case strings.ToLower(correlation.CorrelationIDHeader):
			eventCorrelation.SessionID = string(header.Value)
		case strings.ToLower(correlation.CorrelationIDHeader):
			eventCorrelation.ScenarioName = string(header.Value)
		}
	}
	ctx = correlation.GetContextWithCorrelationParam(ctx, &eventCorrelation)
	return ctx
}
