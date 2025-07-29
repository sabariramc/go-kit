package kafka_test

import (
	"context"
	"net"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/sabariramc/go-kit/env"
	ck "github.com/sabariramc/go-kit/kafka"
	"github.com/sabariramc/go-kit/log"
	"github.com/segmentio/kafka-go"
)

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
			NumPartitions:     10, // Set to your desired number of partitions
			ReplicationFactor: 1,  // Or your desired replication factor
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
	TopicMultiplePartitionsOne   = "TopicMultiplePartitions.One"
	TopicMultiplePartitionsTwo   = "TopicMultiplePartitions.Two"
	TopicMultiplePartitionsThree = "TopicMultiplePartitions.Three"
	TopicMultiplePartitionsFour  = "TopicMultiplePartitions.Four"
)

func TestMain(m *testing.M) {
	os.Setenv("LOG_LEVEL", "trace")
	logger = log.New("KafkaTest.TestMain", log.WithConsole())
	cmdUp := exec.Command("docker", "compose", "up", "-d")
	cmdUp.Stdout = os.Stdout
	cmdUp.Stderr = os.Stderr
	if err := cmdUp.Run(); err != nil {
		logger.Fatal(context.Background()).Err(err).Msg("Failed to bring up Docker Compose")
	}
	time.Sleep(10 * time.Second)
	if err := CreateTopic(TopicMultiplePartitionsOne, TopicMultiplePartitionsTwo, TopicMultiplePartitionsThree, TopicMultiplePartitionsFour); err != nil {
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
