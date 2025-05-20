package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Publish sends a message to Kafka (synchronous with timeout)
func (kc *KafkaClient) Publish(ctx context.Context, topic string, msg Message) error {
	if kc.producer == nil {
		return ErrProducerNotInitialized
	}

	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: msg.Value,
	}

	if msg.Key != "" {
		kafkaMsg.Key = []byte(msg.Key)
	}

	if len(msg.Headers) > 0 {
		headers := make([]kafka.Header, 0, len(msg.Headers))
		for k, v := range msg.Headers {
			headers = append(headers, kafka.Header{
				Key:   k,
				Value: []byte(v),
			})
		}
		kafkaMsg.Headers = headers
	}

	deliveryChan := make(chan kafka.Event)
	defer close(deliveryChan)

	err := kc.producer.Produce(kafkaMsg, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	select {
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			return fmt.Errorf("delivery failed: %w", m.TopicPartition.Error)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(defaultTimeout):
		return errors.New("produce timeout")
	}
}

// PublishAsync sends a message to Kafka asynchronously
func (kc *KafkaClient) PublishAsync(topic string, msg Message) error {
	if kc.producer == nil {
		return ErrProducerNotInitialized
	}

	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: msg.Value,
	}

	if msg.Key != "" {
		kafkaMsg.Key = []byte(msg.Key)
	}

	if len(msg.Headers) > 0 {
		headers := make([]kafka.Header, 0, len(msg.Headers))
		for k, v := range msg.Headers {
			headers = append(headers, kafka.Header{
				Key:   k,
				Value: []byte(v),
			})
		}
		kafkaMsg.Headers = headers
	}

	return kc.producer.Produce(kafkaMsg, nil)
}

// PublishJSON is a convenience method for publishing JSON messages
func (kc *KafkaClient) PublishJSON(ctx context.Context, topic string, key string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	msg := Message{
		Key:   key,
		Value: jsonData,
	}

	return kc.Publish(ctx, topic, msg)
}
