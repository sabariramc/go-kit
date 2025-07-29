package consumer

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	span "github.com/sabariramc/go-kit/instrumentation"

	"github.com/sabariramc/go-kit/log"
	"github.com/segmentio/kafka-go"
)

type Partition struct {
	Topic     string
	Partition int
}

type OffsetMap map[Partition]int64

func (o OffsetMap) Copy() OffsetMap {
	res := make(OffsetMap, len(o))
	for partition, offset := range o {
		res[partition] = offset
	}
	return res
}

func (o OffsetMap) MarshalZerologObject(e *zerolog.Event) {
	for partition, offset := range o {
		e.Int64(partition.Topic+":"+strconv.Itoa(partition.Partition), offset)
	}
}

type topics []string

func (t topics) MarshalZerologArray(a *zerolog.Array) {
	for _, topic := range t {
		a.Str(topic)
	}
}

type MessageWithContext struct {
	*kafka.Message
	Ctx context.Context
}

type AutoCommit struct {
	Enabled      bool
	IntervalInMs uint64
	BatchSize    uint64
}

type Reader struct {
	*kafka.Reader
	log              *log.Logger
	commitLock       sync.Mutex
	consumedOffset   OffsetMap
	committedOffset  OffsetMap
	autoCommit       AutoCommit
	autoCommitCancel context.CancelFunc
	pollCancel       context.CancelFunc
	wg               sync.WaitGroup
	topics           topics
	count            uint64
	pollLock         sync.Mutex
	msgContext       []MessageContext
	tr               span.SpanOp
	closePollCh      bool
}

func New(ctx context.Context, options ...Options) (*Reader, error) {
	config := NewDefaultConfig()
	for _, opt := range options {
		err := opt(config)
		if err != nil {
			return nil, fmt.Errorf("kafka.Reader.New: error applying option: %w", err)
		}
	}
	err := ValidateConfig(config)
	if err != nil {
		return nil, fmt.Errorf("kafka.Reader.New: config validation error: %w", err)
	}
	reader := kafka.NewReader(*config.ReaderConfig)
	k := &Reader{
		Reader:         reader,
		log:            config.Log,
		autoCommit:     config.AutoCommit,
		consumedOffset: make(OffsetMap),
		topics:         config.ReaderConfig.GroupTopics,
		tr:             config.SpanOp,
		msgContext:     config.MessageContext,
		closePollCh:    config.ClosePollCh,
	}
	return k, nil
}

func (k *Reader) Poll(ctx context.Context, ch chan<- *MessageWithContext) (offset OffsetMap, err error) {
	k.pollLock.Lock()
	defer k.pollLock.Unlock()
	if k.closePollCh {
		defer close(ch)
	}
	k.wg.Add(1)
	defer k.wg.Done()
	k.log.Info(ctx).Array("topics", k.topics).Msg("Polling started for topics")
	nCtx := context.WithoutCancel(ctx)
	readerClosed, cancel := context.WithCancel(ctx)
	k.pollCancel = cancel
	go k.autoCommitTimeBased(readerClosed)
	var msg kafka.Message
	var commitErr error
forLoop:
	for {
		msg, err = k.FetchMessage(ctx)
		if err != nil {
			err = fmt.Errorf("error fetching message: %w", err)
			if k.autoCommit.Enabled {
				offset, commitErr = k.Commit(nCtx)
			}
			break
		}
		select {
		case <-ctx.Done():
			offset, commitErr = k.Commit(nCtx)
			err = ctx.Err()
			break forLoop
		case <-readerClosed.Done():
			break forLoop
		case ch <- &MessageWithContext{Message: &msg, Ctx: k.getMessageContext(&msg)}:
			k.storeOffset(&msg)
			offset, commitErr = k.autoCommitSizeBased(nCtx)
			if commitErr != nil {
				break forLoop
			}
		}
	}
	if commitErr != nil || err != nil {
		if err == nil {
			err = fmt.Errorf("kafka.Reader.Poll: commitError: %w", commitErr)
		} else if commitErr != nil {
			err = fmt.Errorf("kafka.Reader.Poll: pollErr: %w , commitError: %w", err, commitErr)
		}
		k.log.Error(ctx).Err(err).Msg("error occurred during polling")
	}
	k.log.Warn(ctx).Array("topics", k.topics).Msg("Polling ended for topics")
	return
}

func (k *Reader) autoCommitSizeBased(ctx context.Context) (OffsetMap, error) {
	if !k.autoCommit.Enabled {
		return nil, nil
	}
	if k.count < k.autoCommit.BatchSize {
		return nil, nil
	}
	k.log.Debug(ctx).Msgf("batch size reached for auto commit, committing messages")
	return k.Commit(ctx)
}

func (k *Reader) autoCommitTimeBased(ctx context.Context) {
	if !k.autoCommit.Enabled {
		k.log.Debug(ctx).Msg("Auto commit is disabled")
		return
	}
	readerClosed, cancel := context.WithCancel(context.Background())
	k.autoCommitCancel = cancel
	k.wg.Add(1)
	timeout, _ := context.WithTimeout(context.Background(), time.Duration(k.autoCommit.IntervalInMs)*time.Millisecond)
	defer k.wg.Done()
	defer k.log.Warn(ctx).Msg("auto commit stopped")
	nCtx := context.WithoutCancel(ctx)
	for {
		select {
		case <-timeout.Done():
			k.log.Debug(ctx).Msgf("time limit reached for auto commit interval, committing messages")
			_, err := k.Commit(nCtx)
			if err != nil {
				k.log.Error(nCtx).Err(err).Msg("Error while writing kafka message")
			}
			timeout, _ = context.WithTimeout(context.Background(), time.Duration(k.autoCommit.IntervalInMs)*time.Millisecond)
		case <-ctx.Done():
			k.log.Debug(ctx).Msg("context cancelled, committing messages")
			_, err := k.Commit(nCtx)
			if err != nil {
				k.log.Error(nCtx).Err(err).Msg("error in auto commit")
			}
			return
		case <-readerClosed.Done():
			k.log.Debug(ctx).Msg("reader closed, committing messages")
			_, err := k.Commit(nCtx)
			if err != nil {
				k.log.Error(nCtx).Err(err).Msg("error in auto commit")
			}
		}
	}
}

func (k *Reader) Close(ctx context.Context) error {
	k.log.Info(ctx).Array("topics", k.topics).Msg("Consumer closer initiated")
	if k.pollCancel != nil {
		k.pollCancel()
	}
	if k.autoCommitCancel != nil {
		k.autoCommitCancel()
	}
	closeErr := k.Reader.Close()
	if closeErr != nil {
		k.log.Error(ctx).Array("topics", k.topics).Err(closeErr).Msg("Consumer closed with error")
		return fmt.Errorf("Consumer.Close: %w", closeErr)
	}
	k.wg.Wait()
	k.log.Info(ctx).Array("topics", k.topics).Msg("Consumer closed successfully")
	return nil
}

func (k *Reader) Commit(ctx context.Context) (OffsetMap, error) {
	k.commitLock.Lock()
	defer k.commitLock.Unlock()
	if len(k.consumedOffset) == 0 {
		k.log.Debug(ctx).Msg("no messages to commit")
		return nil, nil
	}
	if k.tr != nil {
		var crSpan span.Span
		ctx, crSpan = k.tr.NewSpanFromContext(ctx, "kafka.consumer.commit", span.SpanKindConsumer, "")
		defer crSpan.Finish()
	}
	msgList := make([]kafka.Message, 0, len(k.consumedOffset))
	for partition, offset := range k.consumedOffset {
		msgList = append(msgList, kafka.Message{
			Topic:     partition.Topic,
			Partition: partition.Partition,
			Offset:    offset,
		})
	}
	k.log.Debug(ctx).Object("offsets", k.consumedOffset).Uint64("no_of_messages", k.count).Msg("initiating commit")
	err := k.CommitMessages(ctx, msgList...)
	if err != nil {
		return nil, fmt.Errorf("kafka.Reader.Commit: error committing message: %w", err)
	}
	k.committedOffset, k.consumedOffset = k.consumedOffset, make(OffsetMap)
	k.count = 0
	k.log.Info(ctx).Object("offsets", k.committedOffset).Msg("messages committed")
	return k.committedOffset.Copy(), nil
}

func (k *Reader) GetOffsets() (committedOffset OffsetMap, consumedOffset OffsetMap) {
	k.commitLock.Lock()
	defer k.commitLock.Unlock()
	return k.committedOffset.Copy(), k.consumedOffset.Copy()
}

func (k *Reader) storeOffset(msg *kafka.Message) {
	k.commitLock.Lock()
	defer k.commitLock.Unlock()
	k.count++
	k.consumedOffset[Partition{
		Topic:     msg.Topic,
		Partition: msg.Partition,
	}] = msg.Offset
}

func (k *Reader) getMessageContext(msg *kafka.Message) context.Context {
	ctx := context.Background()
	for _, hook := range k.msgContext {
		ctx = hook.Get(ctx, msg)
	}
	return ctx
}
