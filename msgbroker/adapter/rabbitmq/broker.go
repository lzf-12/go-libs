package rabbitmq

import "github.com/streadway/amqp"

type RabbitMQBroker struct {
	conn        *amqp.Connection
	producercfg ProducerCfg
	consumercfg ConsumerCfg
}

type RabbitMQOpts struct {
	AmqpString string
	AmqpConfig amqp.Config
}

type ProducerInt interface {
	Publish(topic string, message []byte) error
	PublishWithHeaders(topic string, message []byte, headers map[string]interface{}) error
}

type ConsumerInt interface {
	Subscribe(topic string, handler func(message []byte, headers map[string]interface{})) error
	Unsubscribe() error
}
