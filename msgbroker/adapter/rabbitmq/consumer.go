package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"msgbroker/retry"
	"sync"
	"testing"

	"github.com/streadway/amqp"
)

type ConsumerInt interface {
	Subscribe(topic string, handler func(message []byte, headers map[string]interface{})) error
	Unsubscribe() error
}

type RabbitMQConsumer struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	config       ConsumerCfg
	done         chan struct{}
	shutdownOnce sync.Once
}

// RabbitMQ consumer-specific configuration
type ConsumerCfg struct {
	QueueName     string
	ConsumerTag   string            // identifier for the consumer
	AutoAck       bool              // automatically acknowledge messages
	Exclusive     bool              // exclusive consumer access
	NoLocal       bool              // don't receive messages published by this connection
	Args          amqp.Table        // additional arguments
	PrefetchCount int               // QoS setting for number of unacknowledged messages
	PrefetchSize  int               // QoS setting for max size of unacknowledged messages
	RetryPolicy   retry.RetryPolicy // For message processing failures
}

func (r *RabbitMQBroker) NewConsumer(config ConsumerCfg) (ConsumerInt, error) {
	channel, err := r.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// set QoS (prefetch count)
	if err := channel.Qos(
		config.PrefetchCount,
		config.PrefetchSize,
		false,
	); err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &RabbitMQConsumer{
		conn:    r.conn,
		channel: channel,
		config:  config,
		done:    make(chan struct{}),
	}, nil
}

func (c *RabbitMQConsumer) Subscribe(topic string, handler func(message []byte, headers map[string]interface{})) error {
	// declare queue if it doesn't exist
	queue, err := c.channel.QueueDeclare(
		c.config.QueueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		c.config.Args,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// bind queue to exchange/topic
	if err := c.channel.QueueBind(
		queue.Name,
		topic,       // routing key
		"amq.topic", // exchange
		false,       // noWait
		nil,         // args
	); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// start consuming messages
	deliveries, err := c.channel.Consume(
		queue.Name,
		c.config.ConsumerTag,
		c.config.AutoAck,
		c.config.Exclusive,
		c.config.NoLocal,
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	// start message processing goroutine
	go c.processMessages(deliveries, handler)

	return nil
}

func (c *RabbitMQConsumer) processMessages(deliveries <-chan amqp.Delivery, handler func([]byte, map[string]interface{})) {
	for {
		select {
		case <-c.done:
			return
		case delivery, ok := <-deliveries:
			if !ok {
				// channel closed, attempt to reconnect
				if err := c.reconnect(); err != nil {
					log.Printf("Failed to reconnect: %v", err)
					return
				}
				continue
			}

			// process message with retry logic
			retry.WithBackoff(c.config.RetryPolicy, func() error {
				handler(delivery.Body, delivery.Headers)

				// manual ack if AutoAck is false
				if !c.config.AutoAck {
					if err := delivery.Ack(false); err != nil {
						return fmt.Errorf("failed to ack message: %w", err)
					}
				}
				return nil
			})
		}
	}
}

func (c *RabbitMQConsumer) reconnect() error {
	channel, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to reopen channel: %w", err)
	}
	c.channel = channel
	return nil
}

func (c *RabbitMQConsumer) Unsubscribe() error {
	c.shutdownOnce.Do(func() {
		close(c.done) // gracefully close multiple go routine
	})
	return nil
}

//
//  EXAMPLE ON HOW TO USE FLEXIBLE HANDLER TO CONSUME AND PROCESS MESSAGE AND METADATA
//  codes below are only example
//

type Order struct {
	OrderId   string `json:"order_id"`
	ProductId string `json:"product_id"`
	Qty       int    `json:"qty"`
}

// example consume order event with body as json
func processOrderJsonHandler(body []byte, headers map[string]interface{}) func(message []byte, headers map[string]interface{}) {
	return func(message []byte, headers map[string]interface{}) {

		//defer saveError
		var order Order
		if err := json.Unmarshal(body, &order); err != nil {
			log.Println("failed processing order event", fmt.Errorf("error marshal: %w", err))
		}

		// handle usecase or repository here
	}
}

func TestSubscribeWithFlexibleHandler(*testing.T) {

	rmq := RabbitMQBroker{}
	consumer, err := rmq.NewConsumer(rmq.consumercfg)
	if err != nil {
		return
	}

	topic := "topic"
	body := []byte{10, 123, 123}
	header := map[string]interface{}{"key": "value"}

	var flexibleHandlerFunc func(message []byte, headers map[string]interface{})
	flexibleHandlerFunc = processOrderJsonHandler(body, header)

	consumer.Subscribe(topic, flexibleHandlerFunc)
}
