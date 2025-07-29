package kafka

const (
	EnvBroker = "KAFKA__BROKER"

	EnvProducerAcknowledge = "KAFKA__PRODUCER__ACKNOWLEDGE"
	EnvProducerAsync       = "KAFKA__PRODUCER__ASYNC"
	EnvProducerTopic       = "KAFKA__PRODUCER__TOPIC"

	EnvConsumerGroupID                = "KAFKA__CONSUMER__GROUP_ID"
	EnvConsumerTopics                 = "KAFKA__CONSUMER__TOPICS"
	EnvConsumerMaxBuffer              = "KAFKA__CONSUMER__MAX_BUFFER"
	EnvConsumerAutoCommit             = "KAFKA__CONSUMER__AUTO_COMMIT"
	EnvConsumerAutoCommitIntervalInMs = "KAFKA__CONSUMER__AUTO_COMMIT_INTERVAL_IN_MS"
	EnvConsumerAutoCommitBatchSize    = "KAFKA__CONSUMER__AUTO_COMMIT_BATCH_SIZE"
)
