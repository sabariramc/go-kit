package kafka_test

import (
	"context"
	"math/rand/v2"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sabariramc/go-kit/kafka/consumer"
	"github.com/sabariramc/go-kit/kafka/producer"
	"github.com/sabariramc/go-kit/log"
	"github.com/sabariramc/go-kit/log/correlation"
	"github.com/segmentio/kafka-go"
	"gotest.tools/v3/assert"
)

func TestKafkaPollSyncWriter(t *testing.T) {
	t.Parallel()
	ctx := correlation.GetContextWithCorrelationParam(context.TODO(), &correlation.EventCorrelation{
		CorrelationID: "TestKafkaPoll" + uuid.NewString(),
		ScenarioID:    "TestKafkaPoll",
	})
	pr, err := producer.New(context.TODO(), func(c *producer.Config) error {
		c.Writer.Async = false
		c.Writer.AllowAutoTopicCreation = true
		return nil
	}, producer.WithLogger(log.New(producer.ModuleProducer, log.WithLogger(&logger.Logger))))
	assert.NilError(t, err)
	defer pr.Close(ctx)
	testKafkaPoll(t, ctx, "TestKafkaPoll", pr, 10)
}

func TestKafkaPollAsyncWriter(t *testing.T) {
	t.Parallel()
	ctx := correlation.GetContextWithCorrelationParam(context.TODO(), &correlation.EventCorrelation{
		CorrelationID: "TestKafkaPollAsyncWriter" + uuid.NewString(),
		ScenarioID:    "TestKafkaPollAsyncWriter",
	})
	pr, err := producer.New(context.TODO(), producer.WithLogger(log.New(producer.ModuleProducer, log.WithLogger(&logger.Logger))))
	assert.NilError(t, err)
	defer pr.Close(ctx)
	testKafkaPoll(t, ctx, TopicMultiplePartitionsTwo, pr, 10000)
}

func TestMultiplePartitions(t *testing.T) {
	t.Parallel()
	ctx := correlation.GetContextWithCorrelationParam(context.TODO(), &correlation.EventCorrelation{
		CorrelationID: "TestMultiplePartitions" + uuid.NewString(),
		ScenarioID:    "TestMultiplePartitions",
	})
	pr, err := producer.New(context.TODO(), producer.WithLogger(log.New(producer.ModuleProducer, log.WithLogger(&logger.Logger))))
	assert.NilError(t, err)
	defer pr.Close(ctx)
	var wg sync.WaitGroup
	ch := make(chan *consumer.MessageWithContext, 100)
	tCtx, cancel := context.WithCancel(ctx)
	consumer := func(ctx context.Context, clientID string) {
		defer wg.Done()
		co, err := consumer.New(context.TODO(), func(c *consumer.Config) error {
			c.ReaderConfig.GroupTopics = []string{TopicMultiplePartitionsThree, TopicMultiplePartitionsFour}
			c.ClosePollCh = false
			c.GroupID = "TestMultiplePartitions"
			return nil
		}, consumer.WithLogger(log.New(consumer.ModuleConsumer, log.WithLogger(&logger.Logger))), consumer.WithClientID(clientID))
		assert.NilError(t, err)
		defer co.Close(context.Background())
		ctx = correlation.GetContextWithCorrelationParam(ctx, &correlation.EventCorrelation{
			CorrelationID: clientID,
		})
		co.Poll(ctx, ch)
	}
	producer := func(topic string) {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			uuid := uuid.NewString()
			ctx := correlation.GetContextWithCorrelationParam(context.TODO(), &correlation.EventCorrelation{
				CorrelationID: "TestMultiplePartitions" + uuid,
			})
			err = pr.WriteMessages(ctx, kafka.Message{
				Topic: topic,
				Key:   []byte(uuid),
				Value: []byte(strconv.Itoa(i)),
			})
			assert.NilError(t, err)
		}
	}
	wg.Add(12)
	count := 0
	go func() {
		defer wg.Done()
		for i := range ch {
			count++
			if count == 600 {
				cancel()
				return
			}
			m := i.Message
			logger.Info(i.Ctx).Msgf("Received topic: %s from partition: %d, key: %s", m.Topic, m.Partition, string(m.Key))
		}
	}()
	go producer(TopicMultiplePartitionsThree)
	go producer(TopicMultiplePartitionsFour)
	time.Sleep(time.Second)
	clientPrefix := "TestMultiplePartitionsConsumer"
	go consumer(tCtx, clientPrefix+"-"+"1")
	time.Sleep(time.Second * 5)
	for i := 0; i < 3; i++ {
		tCtx, _ := context.WithTimeout(ctx, time.Second*time.Duration(rand.IntN(5)+5))
		go consumer(tCtx, clientPrefix+"-"+strconv.Itoa(i+2))
	}
	go producer(TopicMultiplePartitionsThree)
	go producer(TopicMultiplePartitionsFour)
	time.Sleep(time.Second * 10)
	go producer(TopicMultiplePartitionsThree)
	go producer(TopicMultiplePartitionsFour)
	go consumer(tCtx, clientPrefix+"-"+"5")
	wg.Wait()
	close(ch)
	logger.Info(ctx).Msgf("Total messages received: %d", count)
	assert.Equal(t, count, 600)
}

func TestKafkaPollWithDelay(t *testing.T) {
	t.Parallel()
	totalCount := 50
	sendCount := 0
	ctx := correlation.GetContextWithCorrelationParam(context.TODO(), &correlation.EventCorrelation{
		CorrelationID: "TestKafkaPollWithDelay" + uuid.NewString(),
		ScenarioID:    "TestKafkaPollWithDelay",
	})
	topic := TopicMultiplePartitionsOne
	pr, err := producer.New(context.TODO(), producer.WithLogger(log.New(producer.ModuleProducer, log.WithLogger(&logger.Logger))))
	assert.NilError(t, err)
	defer pr.Close(ctx)
	ch := make(chan *consumer.MessageWithContext, 100)
	msgCount := 0
	count := 0
	uuidVal := "TestKafkaPollWithDelay" + uuid.NewString()
	go func() {
		for i := range ch {
			m := i.Message
			if string(m.Key) == uuidVal {
				msgCount++
			}
			count++
		}
	}()
	for i := 0; i < 10; i++ {
		err = pr.WriteMessages(ctx, kafka.Message{
			Topic: topic,
			Key:   []byte(uuidVal),
			Value: []byte(strconv.Itoa(i) + "-" + uuidVal),
		})

		assert.NilError(t, err)
		sendCount++
	}
	newConsumer := func() *consumer.Reader {
		co, err := consumer.New(context.TODO(), func(c *consumer.Config) error {
			c.ReaderConfig.GroupTopics = []string{topic}
			c.ReaderConfig.GroupID = "TestKafkaPollWithDelay"
			return nil
		}, consumer.WithLogger(log.New(consumer.ModuleConsumer, log.WithLogger(&logger.Logger))))
		assert.NilError(t, err)
		return co
	}
	co := newConsumer()
	go co.Poll(ctx, ch)
	time.Sleep(time.Second * 5)
	err = co.Close(ctx)
	assert.NilError(t, err)
	co = newConsumer()
	assert.NilError(t, err)
	defer co.Close(ctx)
	ch = make(chan *consumer.MessageWithContext, 100)
	tCtx, cancel := context.WithCancel(ctx)
	go co.Poll(tCtx, ch)
	go func() {
		for i := sendCount; i < totalCount; i++ {
			err = pr.WriteMessages(ctx, kafka.Message{
				Topic: topic,
				Key:   []byte(uuidVal),
				Value: []byte(strconv.Itoa(i) + "-" + uuidVal),
			})
		}
	}()
	for i := range ch {
		m := i.Message
		if string(m.Key) == uuidVal {
			msgCount++
		}
		if totalCount == msgCount {
			cancel()
		}
		count++
	}
	logger.Info(ctx).Msgf("Total matched: %d\n", msgCount)
	logger.Info(ctx).Msgf("Total received: %d\n", count)
	assert.Equal(t, totalCount, count)
}

func testKafkaPoll(t *testing.T, ctx context.Context, topic string, pr *producer.Writer, totalCount int) {
	var wg sync.WaitGroup
	wg.Add(1)
	uuidVal := "TestKafkaPoll" + uuid.NewString()
	go func() {
		defer wg.Done()
		for i := 0; i < totalCount; i++ {
			err := pr.WriteMessages(ctx, kafka.Message{
				Topic: topic,
				Key:   []byte(uuidVal),
				Value: []byte(strconv.Itoa(i) + "-" + uuidVal),
			})
			assert.NilError(t, err)
		}
	}()
	time.Sleep(time.Second * 3)
	co, err := consumer.New(ctx, func(c *consumer.Config) error {
		c.ReaderConfig.GroupTopics = []string{topic}
		c.ReaderConfig.GroupID = topic
		return nil
	}, consumer.WithLogger(log.New(consumer.ModuleConsumer, log.WithLogger(&logger.Logger))))
	assert.NilError(t, err)
	defer co.Close(ctx)
	tCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	st := time.Now()
	ch := make(chan *consumer.MessageWithContext, 100)
	go co.Poll(tCtx, ch)
	count := 0
	msgCount := 0
	for i := range ch {
		m := i.Message
		if string(m.Key) == uuidVal {
			count++
		}
		if totalCount == count {
			cancel()
		}
		msgCount++
	}
	logger.Info(ctx).Msgf("Total matched: %d", count)
	logger.Info(ctx).Msgf("Total received: %d", msgCount)
	wg.Wait()
	logger.Info(ctx).Msgf("Time taken in ms: %d", time.Since(st)/1000000)
	assert.Equal(t, totalCount, count)
}
