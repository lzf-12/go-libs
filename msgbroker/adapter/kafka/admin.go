package kafka

import (
	"context"
	"errors"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// CreateTopic creates a new Kafka topic
func (kc *KafkaClient) CreateTopic(ctx context.Context, topic string, partitions int, replicationFactor int) error {
	if kc.adminClient == nil {
		return errors.New("admin client not initialized")
	}

	results, err := kc.adminClient.CreateTopics(
		ctx,
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: replicationFactor,
		}},
		kafka.SetAdminOperationTimeout(defaultTimeout),
	)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError {
			return fmt.Errorf("topic creation failed: %v", result.Error)
		}
	}

	return nil
}
