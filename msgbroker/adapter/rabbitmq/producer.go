package rabbitmq

import (
	"errors"
	"fmt"
	"msgbroker/retry"
	"time"

	"github.com/streadway/amqp"
)

type RabbitMQProducer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  ProducerCfg
}

// RabbitMQ producer-specific configuration
type ProducerCfg struct {
	Exchange      string
	RoutingKey    string
	Mandatory     bool              // Return an error if message can't be routed
	Immediate     bool              // Return an error if message can't be delivered immediately
	ContentType   string            // e.g., "application/json"
	DeliveryMode  uint8             // 1=non-persistent, 2=persistent
	Headers       amqp.Table        // additional headers
	RetryPolicy   retry.RetryPolicy // retry configuration when producer failed publish message
	IsNeedConfirm bool              // default false
}

func NewRabbitMQBroker(opts RabbitMQOpts) (*RabbitMQBroker, error) {
	conn, err := amqp.DialConfig(opts.AmqpString, opts.AmqpConfig)
	if err != nil {
		return nil, err
	}
	return &RabbitMQBroker{conn: conn}, nil
}

func (r *RabbitMQBroker) NewProducer(cfg ProducerCfg) (ProducerInt, error) {

	channel, err := r.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	if err := channel.Confirm(cfg.IsNeedConfirm); err != nil {
		return nil, fmt.Errorf("failed to put channel in confirm mode: %w", err)
	}

	return &RabbitMQProducer{
		conn:    r.conn,
		channel: channel,
		config:  r.producercfg,
	}, nil
}

func (p *RabbitMQProducer) Publish(topic string, message []byte) error {
	return p.PublishWithHeaders(topic, message, nil)
}

func (p *RabbitMQProducer) PublishWithHeaders(topic string, message []byte, headers map[string]interface{}) error {
	routingKey := p.config.RoutingKey
	if routingKey == "" {
		routingKey = topic
	}

	msg := amqp.Publishing{
		DeliveryMode: p.config.DeliveryMode,
		ContentType:  p.config.ContentType,
		Body:         message,
		Headers:      amqp.Table(headers),
		Timestamp:    time.Now(),
	}

	var lastErr error

	retry.WithBackoff(p.config.RetryPolicy, func() error {
		err := p.channel.Publish(
			p.config.Exchange,
			routingKey,
			p.config.Mandatory,
			p.config.Immediate,
			msg,
		)
		if err != nil {
			lastErr = err

			// check if connection/channel needs to be re-connect and retry based on retry policy
			if errors.Is(err, amqp.ErrClosed) {
				if err := p.reconnect(); err != nil {
					return fmt.Errorf("reconnect failed: %w", err)
				}
			}
			return err
		}
		return nil
	})

	return lastErr
}

func (p *RabbitMQProducer) reconnect() error {
	channel, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to reopen channel: %w", err)
	}
	p.channel = channel
	return nil
}

func (p *RabbitMQProducer) Close() error {
	return p.channel.Close()
}
