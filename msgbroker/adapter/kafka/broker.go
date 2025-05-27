package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	defaultTimeout = 10 * time.Second
)

var (
	ErrProducerNotInitialized = errors.New("producer not initialized")
	ErrConsumerNotInitialized = errors.New("consumer not initialized")
)

type Message struct {
	Key       string
	Value     []byte
	Headers   map[string]string
	Timestamp time.Time
	Topic     *string
}

type KafkaClient struct {
	Producer     *kafka.Producer
	Consumer     *kafka.Consumer
	Admin        *kafka.AdminClient
	ConfigMap    *kafka.ConfigMap
	errorChannel chan error
}

func NewKafkaConfigMap() *kafka.ConfigMap {
	return &kafka.ConfigMap{}
}

func NewKafkaConsumerClient(consumerCfgMap *kafka.ConfigMap, adminCfgMap *kafka.ConfigMap) (*KafkaClient, error) {

	// set some sane defaults if not provided
	if _, ok := (*consumerCfgMap)["go.events.channel.enable"]; !ok {
		(*consumerCfgMap)["go.events.channel.enable"] = true
	}

	if _, ok := (*consumerCfgMap)["enable.auto.commit"]; !ok {
		(*consumerCfgMap)["enable.auto.commit"] = false // Prefer manual commits for reliability
	}

	var consumerClient *kafka.Consumer

	consumerClient, err := kafka.NewConsumer(consumerCfgMap)
	if err != nil {
		log.Println("error new consumer: ", err)
		consumerClient.Close()
		return nil, fmt.Errorf("failed to create new consumer: %w", err)
	}

	// derive admin client from consumer
	adminFromClient, err := kafka.NewAdminClientFromConsumer(consumerClient)
	if err != nil {
		log.Println("error new admin from consumer: ", err)
		adminFromClient.Close()
		return nil, fmt.Errorf("failed to create new admin from consumer: %w", err)
	}

	client := &KafkaClient{
		Admin:        adminFromClient,
		Consumer:     consumerClient,
		ConfigMap:    consumerCfgMap,
		errorChannel: make(chan error, 10), // Buffered channel to avoid blocking
	}

	return client, nil
}

func NewKafkaProducerClient(producerCfgMap *kafka.ConfigMap) (*KafkaClient, error) {

	// set some sane defaults if not provided
	if _, ok := (*producerCfgMap)["go.events.channel.enable"]; !ok {
		(*producerCfgMap)["go.events.channel.enable"] = true
	}

	if _, ok := (*producerCfgMap)["enable.auto.commit"]; !ok {
		(*producerCfgMap)["enable.auto.commit"] = false
	}

	var producerClient *kafka.Producer

	producerClient, err := kafka.NewProducer(producerCfgMap)
	if err != nil {
		producerClient.Close()
		return nil, fmt.Errorf("failed to create new consumer: %w", err)
	}

	client := &KafkaClient{
		Producer:     producerClient,
		ConfigMap:    producerCfgMap,
		errorChannel: make(chan error, 10), // Buffered channel to avoid blocking
	}

	// start goroutine to handle delivery reports and errors
	go client.handleEvents()

	return client, nil
}

// handleEvents processes delivery reports and errors from producer
func (kc *KafkaClient) handleEvents() {
	for {
		select {
		case e := <-kc.Producer.Events():
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					kc.errorChannel <- fmt.Errorf("delivery failed: %v", ev.TopicPartition.Error)
				}
			case kafka.Error:
				kc.errorChannel <- fmt.Errorf("producer error: %v", ev)
			}
		}
	}
}

// TODO: need to find way to cleaner or simpler implementation
// close gracefully shuts down the all type kafka client
func (kc *KafkaClient) Close() error {
	var errs []error

	if kc.Producer != nil {

		// flushing for message guarantee, prevent loss, and orderly shutdown
		remainingevents := kc.Producer.Flush(int(defaultTimeout.Milliseconds()))

		if remainingevents > 0 {

			log.Printf("Warning: %d messages remain after initial flush, retrying again ...", remainingevents)
			undelivered := kc.getUndeliveredMessages(remainingevents)

			// Log the undelivered messages with details
			kc.logUndeliveredMessages(undelivered)

			// store error about undelivered messages
			errs = append(errs, fmt.Errorf("%d messages were not delivered", remainingevents))
		}

		kc.Producer.Close()
	}

	if kc.Consumer != nil {
		if err := kc.Consumer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close consumer: %w", err))
		}
	}

	if kc.Admin != nil {
		kc.Admin.Close()
	}

	close(kc.errorChannel)

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	return nil
}

// ErrorChannel returns a channel for receiving asynchronous errors
func (kc *KafkaClient) ErrorChannel() <-chan error {
	return kc.errorChannel
}

// HealthCheck verifies Kafka connectivity
func (kc *KafkaClient) HealthCheck(ctx context.Context) error {
	if kc.Admin == nil {
		return errors.New("admin client not initialized")
	}

	_, err := kc.Admin.GetMetadata(nil, true, int(defaultTimeout.Milliseconds()))
	if err != nil {
		return fmt.Errorf("kafka health check failed: %w", err)
	}

	return nil
}

// getUndeliveredMessages retrieves messages still in the producer queue
func (kc *KafkaClient) getUndeliveredMessages(count int) []*kafka.Message {
	undelivered := make([]*kafka.Message, 0, count)

	// Drain the event channel to check for undelivered messages
	for len(undelivered) < count {
		select {
		case ev := <-kc.Producer.Events():
			if msg, ok := ev.(*kafka.Message); ok {
				if msg.TopicPartition.Error != nil {
					undelivered = append(undelivered, msg)
				}
			}
		default:
			// No more events immediately available
			return undelivered
		}
	}

	return undelivered
}

// logUndeliveredMessages logs details about failed messages
func (kc *KafkaClient) logUndeliveredMessages(messages []*kafka.Message) {
	if len(messages) == 0 {
		return
	}

	log.Printf("--- UNDELIVERED MESSAGES REPORT ---")
	log.Printf("Count: %d", len(messages))

	for i, msg := range messages {
		entry := struct {
			Index     int
			Topic     string
			Partition int32
			Key       string
			Value     string
			Error     string
		}{
			Index:     i + 1,
			Topic:     *msg.TopicPartition.Topic,
			Partition: msg.TopicPartition.Partition,
			Error:     msg.TopicPartition.Error.Error(),
		}

		if msg.Key != nil {
			entry.Key = string(msg.Key)
		}

		if msg.Value != nil {
			entry.Value = string(msg.Value)
		}

		// Using JSON marshaling for structured logging
		jsonData, _ := json.Marshal(entry)
		log.Printf("Undelivered message: %s", jsonData)
	}

	log.Printf("--- END UNDELIVERED MESSAGES ---")
}
