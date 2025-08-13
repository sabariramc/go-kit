package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/sabariramc/go-kit/app/base"
	"github.com/sabariramc/go-kit/kafka/consumer"
	"github.com/sabariramc/go-kit/log"
	"github.com/sabariramc/go-kit/log/correlation"
	"github.com/segmentio/kafka-go"
)

type Handler interface {
	Handle(context.Context, *kafka.Message) error
}

type HandlerFunc func(context.Context, *kafka.Message) error

func (f HandlerFunc) Handle(ctx context.Context, msg *kafka.Message) error {
	return f(ctx, msg)
}

type KafkaConsumer struct {
	*consumer.Reader
	*base.Base
	log        *log.Logger
	ch         chan *consumer.MessageWithContext
	handler    map[string]Handler
	tr         Tracer
	stop       context.CancelFunc
	shutdownWG sync.WaitGroup
	topics     map[string]struct{}
}

// New creates a new instance of kafka.
func New(option ...Options) (*KafkaConsumer, error) {
	cfg, err := NewConfig()
	if err != nil {
		return nil, err
	}
	k := &KafkaConsumer{
		Reader:  cfg.Reader,
		Base:    cfg.Base,
		log:     cfg.Log,
		handler: make(map[string]Handler),
		tr:      cfg.Tracer,
		ch:      make(chan *consumer.MessageWithContext, cfg.MessageChannelSize),
		topics:  make(map[string]struct{}),
	}
	for _, val := range k.GetTopics() {
		k.topics[val] = struct{}{}
	}
	k.RegisterHooks(k)
	k.shutdownWG.Add(1)
	return k, nil
}

func (k *KafkaConsumer) Close(ctx context.Context) error {
	k.stop()
	k.shutdownWG.Wait()
	k.Reader.Close(ctx)
	return nil
}

func (k *KafkaConsumer) Start(ctx context.Context) {
	corr := &correlation.EventCorrelation{CorrelationID: fmt.Sprintf("%v:kafka", base.GetServiceName())}
	stopCtx, stop := context.WithCancel(correlation.GetContextWithCorrelationParam(ctx, corr))
	k.stop = stop
	defer k.AwaitShutdownCompletion()
	defer k.shutdownWG.Done()
	shutdownCtx := correlation.GetContextWithCorrelationParam(context.Background(), &correlation.EventCorrelation{
		CorrelationID: "KafkaConsumerShutdown",
	})
	k.StartSignalMonitor(shutdownCtx)
	var pollWg sync.WaitGroup
	defer pollWg.Wait()
	pollWg.Add(1)
	pollCtx, cancelPoll := context.WithCancel(correlation.GetContextWithCorrelationParam(context.Background(), corr))
	go func() {
		defer pollWg.Done()
		offset, err := k.Poll(pollCtx, k.ch)
		if err != nil && !errors.Is(err, context.Canceled) {
			k.log.Error(stopCtx).Err(err).Object("offsets", offset).Msg("Kafka consumer exited")
		}
		go k.Shutdown(shutdownCtx)
	}()
	k.log.Info(stopCtx).Msg("kafka consumer started")
	defer k.log.Info(stopCtx).Msg("kafka consumer stopped")
	for {
		select {
		case <-stopCtx.Done():
			cancelPoll()
			return
		case msg, ok := <-k.ch:
			if !ok {
				return
			}
			k.Handle(msg.Ctx, msg.Message)
		}
	}
}
