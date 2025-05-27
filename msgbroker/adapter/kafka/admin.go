package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// create topic only if not exist, otherwise skip creating topic
func (kc *KafkaClient) CreateTopicsIfNotExist(ctx context.Context, topicHandlers []TopicHandler) error {
	if kc.Admin == nil {
		return errors.New("admin client not initialized")
	}

	for _, th := range topicHandlers {

		// first, check if the topic already exists
		topicDetails, err := kc.Admin.GetMetadata(&th.Topic, false, int(defaultTimeout.Seconds()))
		if err != nil {
			return fmt.Errorf("failed to get topic metadata: %w", err)
		}

		// if the topic exists in metadata, check if has error unknown or not
		if tp, exists := topicDetails.Topics[th.Topic]; exists {
			if tp.Error.Code() == kafka.ErrUnknownTopic || tp.Error.Code() == kafka.ErrUnknownTopicOrPart {

				// create topic
				results, err := kc.Admin.CreateTopics(
					ctx,
					[]kafka.TopicSpecification{{
						Topic:             th.Topic,
						NumPartitions:     th.Partitions,
						ReplicationFactor: th.ReplicationFactor,
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

			// topic exist
			if tp.Error.Code() == kafka.ErrNoError {
				log.Printf("topic %s already exist, skip creating topic...", th.Topic)
				continue
			}
		}

	}

	return nil
}
