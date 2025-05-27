package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type TopicHandler struct {
	Topic             string
	Handler           func(Message) error
	Partitions        int
	ReplicationFactor int
}

// subscribe starts consuming messages from a topic
func (kc *KafkaClient) SubscribeTopics(ctx context.Context, topicHandlers []TopicHandler) error {
	if kc.Consumer == nil {
		return ErrConsumerNotInitialized
	}

	if len(topicHandlers) < 1 {
		return errors.New("error topic and handler map cannot empty")
	}

	handlerMap := make(map[string]func(Message) error)
	var topics []string
	for _, th := range topicHandlers {
		handlerMap[th.Topic] = th.Handler
		topics = append(topics, th.Topic)
	}

	cb := func(consumer *kafka.Consumer, event kafka.Event) error {

		log.Println("received new callback event: ", event.String())
		switch ev := event.(type) {
		case kafka.AssignedPartitions:
			consumer.Assign(ev.Partitions)

			for _, p := range ev.Partitions {
				log.Printf("assigned topic=%s partition=%v offset=%v", getTopicName(p.Topic), p.Partition, p.Offset)
			}

		case kafka.RevokedPartitions:
			consumer.Commit()
			for _, p := range ev.Partitions {
				log.Printf("revoked topic=%s partition=%v offset=%v", getTopicName(p.Topic), p.Partition, p.Offset)
			}
		}
		return nil
	}

	err := kc.Consumer.SubscribeTopics(topics, cb)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("unsub consumers from all topic...")
			kc.Consumer.Unsubscribe()
			log.Println("unsub done")
			time.Sleep(5 * time.Second)
			return nil
		default:
			msg, err := kc.Consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				return fmt.Errorf("consumer error: %w", err)
			}

			message := Message{
				Topic:     msg.TopicPartition.Topic,
				Value:     msg.Value,
				Timestamp: msg.Timestamp,
			}

			if msg.Key != nil {
				message.Key = string(msg.Key)
			}

			if len(msg.Headers) > 0 {
				headers := make(map[string]string)
				for _, header := range msg.Headers {
					headers[header.Key] = string(header.Value)
				}
				message.Headers = headers
			}

			// get the appropriate handler for this topic
			handler, exists := handlerMap[*msg.TopicPartition.Topic]
			if !exists {
				log.Printf("no handler found for topic: %s", *msg.TopicPartition.Topic)
				continue
			}

			if err := handler(message); err != nil {
				log.Printf("message handling failed for topic %s: %v", *msg.TopicPartition.Topic, err)
				continue
			}

			// manual commit after successful processing
			_, err = kc.Consumer.CommitMessage(msg)
			if err != nil {
				log.Printf("failed to commit message: %v", err)
			}
		}
	}
}

func getTopicName(topic *string) string {
	if topic == nil {
		return "nil"
	}
	return *topic
}
