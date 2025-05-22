package rabbitmq

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/lzf-12/go-example-collections/msgbroker/retry"

	"github.com/streadway/amqp"
)

type ConsumerInt interface {
	Subscribe(queuename string, topic string, handler func(message []byte, headers map[string]interface{})) error
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
	DLQExchange   string
	DLQRoutingKey string
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

func (c *RabbitMQConsumer) Subscribe(queuename string, topic string, handler func(message []byte, headers map[string]interface{})) error {
	// declare queue if it doesn't exist
	queue, err := c.channel.QueueDeclare(
		queuename,
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
		queuename,
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
			err := retry.WithBackoff(c.config.RetryPolicy, func() error {
				handler(delivery.Body, delivery.Headers)

				// manual ack if AutoAck is false
				if !c.config.AutoAck {
					if err := delivery.Ack(false); err != nil {
						return fmt.Errorf("failed to ack message: %w", err)
					}
				}
				return nil
			})

			// if max retries failed, move to DLQ
			if err != nil {
				log.Printf("Handler failed after max retries, sending to DLQ: %v", err)
				c.sendToDLQ(delivery)
			}
		}
	}
}

func (c *RabbitMQConsumer) sendToDLQ(delivery amqp.Delivery) {

	dlqExchange := c.config.DLQExchange     // default configuration
	dlqRoutingKey := c.config.DLQRoutingKey // default configuration

	err := c.channel.Publish(
		dlqExchange,   // exchange
		dlqRoutingKey, // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType:   delivery.ContentType,
			Body:          delivery.Body,
			Headers:       delivery.Headers,
			Timestamp:     time.Now(),
			CorrelationId: delivery.CorrelationId,
			DeliveryMode:  delivery.DeliveryMode,
			MessageId:     delivery.MessageId,
			Type:          delivery.Type,
			AppId:         delivery.AppId,
		},
	)

	if err != nil {
		log.Printf("Failed to send message to DLQ: %v", err)
	} else {
		// Always ack or reject the original message to avoid requeue
		if err := delivery.Ack(false); err != nil {
			log.Printf("Failed to ack original message after DLQ forward: %v", err)
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
