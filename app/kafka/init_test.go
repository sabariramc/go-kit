package kafka_test

import (
	"context"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	consumer "github.com/sabariramc/go-kit/app/kafka"
	"github.com/sabariramc/go-kit/env"
	ck "github.com/sabariramc/go-kit/kafka"
	"github.com/sabariramc/go-kit/log"
	"github.com/segmentio/kafka-go"
	"gotest.tools/v3/assert"
)

type TestConsumer struct {
	*consumer.KafkaConsumer
	ch chan *kafka.Message
}

func new() (*TestConsumer, error) {
	k, err := consumer.New()
	if err != nil {
		return nil, err
	}
	tc := &TestConsumer{
		KafkaConsumer: k,
		ch:            make(chan *kafka.Message, 10),
	}
	tc.registerHandlers()
	return tc, nil
}

func (t *TestConsumer) registerHandlers() {
	t.KafkaConsumer.AddHandler(context.TODO(), TopicOne, consumer.HandlerFunc(func(ctx context.Context, msg *kafka.Message) error {
		t.ch <- msg
		return nil
	}))
	t.KafkaConsumer.AddHandler(context.TODO(), TopicTwo, consumer.HandlerFunc(func(ctx context.Context, msg *kafka.Message) error {
		t.ch <- msg
		return nil
	}))
}

func New(t *testing.T) *TestConsumer {
	tc, err := new()
	assert.NilError(t, err, "Failed to create new TestConsumer")
	return tc
}

func CreateTopic(topics ...string) error {

	// Connect to Kafka broker
	broker := env.GetSlice(ck.EnvBroker, []string{"localhost:9093"}, ",")
	conn, err := kafka.Dial("tcp", broker[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	// Get the controller node information
	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	// Define topic configuration
	topicConfigs := []kafka.TopicConfig{}
	for _, topic := range topics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     5, // Set to your desired number of partitions
			ReplicationFactor: 1, // Or your desired replication factor
		})
	}

	// Create topic
	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return err
	}
	return nil
}

var logger *log.Logger

const (
	TopicOne   = "TopicMultiplePartitions.One"
	TopicTwo   = "TopicMultiplePartitions.Two"
	TopicThree = "TopicMultiplePartitions.Three"
)

func TestMain(m *testing.M) {
	os.Setenv(log.EnvLogLevel, "trace")
	os.Setenv(log.EnvLogFormat, "console")
	os.Setenv(ck.EnvConsumerTopics, strings.Join([]string{TopicOne, TopicTwo, TopicThree}, ","))
	logger = log.New("KafkaTest.TestMain")
	cmdUp := exec.Command("docker", "compose", "up", "-d")
	cmdUp.Stdout = os.Stdout
	cmdUp.Stderr = os.Stderr
	if err := cmdUp.Run(); err != nil {
		logger.Fatal(context.Background()).Err(err).Msg("Failed to bring up Docker Compose")
	}
	time.Sleep(10 * time.Second)
	if err := CreateTopic(TopicOne, TopicTwo, TopicThree); err != nil {
		logger.Fatal(context.Background()).Err(err).Msg("Failed to create Kafka topic")
	}

	code := m.Run()

	cmdDown := exec.Command("docker", "compose", "down")
	cmdDown.Stdout = os.Stdout
	cmdDown.Stderr = os.Stderr
	if err := cmdDown.Run(); err != nil {
		logger.Error(context.Background()).Err(err).Msg("Failed to bring down Docker Compose")
	}
	time.Sleep(10 * time.Second)
	os.Exit(code)
}
