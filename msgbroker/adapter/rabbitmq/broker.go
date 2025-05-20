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

func NewRabbitMQBroker(opts RabbitMQOpts) (*RabbitMQBroker, error) {
	conn, err := amqp.DialConfig(opts.AmqpString, opts.AmqpConfig)
	if err != nil {
		return nil, err
	}
	return &RabbitMQBroker{conn: conn}, nil
}
