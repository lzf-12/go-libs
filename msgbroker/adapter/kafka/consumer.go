package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Subscribe starts consuming messages from a topic
func (kc *KafkaClient) Subscribe(ctx context.Context, topics []string, handler func(Message) error) error {
	if kc.consumer == nil {
		return ErrConsumerNotInitialized
	}

	err := kc.consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := kc.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				return fmt.Errorf("consumer error: %w", err)
			}

			message := Message{
				Value: msg.Value,
			}

			if msg.Key != nil {
				message.Key = string(msg.Key)
			}

			message.Timestamp = msg.Timestamp

			if len(msg.Headers) > 0 {
				headers := make(map[string]string)
				for _, header := range msg.Headers {
					headers[header.Key] = string(header.Value)
				}
				message.Headers = headers
			}

			if err := handler(message); err != nil {
				log.Printf("message handling failed: %v", err)
				continue
			}

			// Manual commit after successful processing
			_, err = kc.consumer.CommitMessage(msg)
			if err != nil {
				log.Printf("failed to commit message: %v", err)
			}
		}
	}
}
