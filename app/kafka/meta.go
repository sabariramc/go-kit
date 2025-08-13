package kafka

import (
	"context"
)

func (k *KafkaConsumer) HealthCheck(ctx context.Context) error {
	k.Reader.Stats()
	return nil
}
