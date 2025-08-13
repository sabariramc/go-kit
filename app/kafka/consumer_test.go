package kafka_test

import (
	"context"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"
	consumer "github.com/sabariramc/go-kit/app/kafka"
	ck "github.com/sabariramc/go-kit/kafka"
	"github.com/sabariramc/go-kit/kafka/producer"
	"github.com/segmentio/kafka-go"
	"gotest.tools/v3/assert"
)

func TestKafkaConsumerSignal(t *testing.T) {
	os.Setenv(ck.EnvConsumerTopics, strings.Join([]string{TopicOne, TopicTwo}, ","))
	consumer := New(t)
	pr, err := producer.New(context.TODO())
	assert.NilError(t, err)
	go func() {
		// Produce test messages
		for i := 0; i < 10; i++ {
			pr.WriteMessages(context.Background(), kafka.Message{
				Topic: TopicOne,
				Key:   []byte(uuid.NewString()),
				Value: []byte(strconv.Itoa(i)),
			})
		}
		for i := 0; i < 5; i++ {
			pr.WriteMessages(context.Background(), kafka.Message{
				Topic: TopicTwo,
				Key:   []byte(uuid.NewString()),
				Value: []byte(strconv.Itoa(i)),
			})
		}
	}()
	go func() {
		count := 0
		for _ = range consumer.ch {
			count++
			if count == 10 {
				break
			}
		}
		// Get the current process and send SIGINT to it.
		proc, err := os.FindProcess(os.Getpid())
		if err != nil {
			panic(err)
		}
		// Send SIGINT (simulates Ctrl+C)
		proc.Signal(syscall.SIGINT)
	}()
	consumer.Start(context.Background())
}

func TestKafkaConsumerContext(t *testing.T) {
	kc := New(t)
	kc.AddHandler(context.Background(), TopicThree, consumer.HandlerFunc(func(ctx context.Context, msg *kafka.Message) error {
		// Handle messages from TopicThree
		return nil
	}))
	pr, err := producer.New(context.TODO())
	assert.NilError(t, err)
	timerCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	go func() {
		for i := 0; ; i++ {
			select {
			case <-timerCtx.Done():
				return
			default:
				pr.WriteMessages(context.Background(), kafka.Message{
					Topic: TopicThree,
					Key:   []byte(uuid.NewString()),
					Value: []byte(strconv.Itoa(i)),
				})
			}
		}
	}()
	go func() {
		for _ = range kc.ch {
		}
	}()
	kc.Start(timerCtx)
}
