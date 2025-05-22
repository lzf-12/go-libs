package rabbitmq

import (
	"log"

	"github.com/streadway/amqp"
)

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

func gracefulShutdown(conn *amqp.Connection, ch *amqp.Channel, cleanup func()) {

	// Call optional cleanup logic
	if cleanup != nil {
		cleanup()
	}

	// Gracefully close channel and connection
	log.Println("Closing RabbitMQ channel and connection...")
	if ch != nil {
		if err := ch.Close(); err != nil {
			log.Printf("Error closing channel: %v\n", err)
		}
	}

	if conn != nil {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v\n", err)
		}
	}

	log.Println("RabbitMQ shutdown complete.")
}
