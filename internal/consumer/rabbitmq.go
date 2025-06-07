package consumer

import (
	"context"
	"log"
	"time"

	"github.com/lzf-12/go-example-collections/internal/config"
	"github.com/lzf-12/go-example-collections/internal/consumer/handler"
	"github.com/lzf-12/go-example-collections/internal/consumer/model"
	"github.com/lzf-12/go-example-collections/msgbroker/adapter/rabbitmq"
	"github.com/lzf-12/go-example-collections/msgbroker/retry"
)

func InitRabbitMQConsumer(ctx context.Context) error {

	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Printf("load config failed: %v", err)
		return err
	}

	opts := rabbitmq.RabbitMQOpts{
		AmqpString: cfg.RabbitMQAmqpString,
	}

	rmq, err := rabbitmq.NewRabbitMQBroker(opts)
	if err != nil {
		log.Printf("rabbitMQ initialize connection failed: %v", err)
		return err
	}

	consumerCfg := rabbitmq.ConsumerCfg{
		PrefetchCount: 0,
		PrefetchSize:  0,
		RetryPolicy: retry.RetryPolicy{
			MaxRetries:      3,
			InitialInterval: 2 * time.Second,
			Multiplier:      2,
			MaxInterval:     10 * time.Second,
		},
		DLQExchange:   cfg.RabbitMQDefaultDlx,
		DLQRoutingKey: cfg.RabbitMQDefaultDlq,
	}

	consumer, err := rmq.NewConsumer(consumerCfg)
	if err != nil {
		log.Printf("consumer initialize failed: %v", err)
		return err
	}

	mapQueueTopicHandler := []model.QueueTopicHandler{
		{
			Queue:   model.RmqQueueOrder,
			Topic:   model.TopicOrderV1Json,
			Handler: handler.HandleCreateOrderV1JSON,
		},
		{
			Queue:   model.RmqQueueOrder,
			Topic:   model.TopicOrderV1Xml,
			Handler: handler.HandleCreateOrderV1XML,
		},
	}

	// subscribe each map
	for _, qth := range mapQueueTopicHandler {
		err := consumer.Subscribe(qth.Queue, qth.Topic, qth.Handler)
		if err != nil {
			log.Printf("failed to subscribe to topic %s: %v", qth.Topic, err)
		} else {
			log.Printf("subscribed to topic: %s", qth.Topic)
		}
	}

	// shutdown context received
	<-ctx.Done()

	log.Println("shutdown signal received in RabbitMQ consumer. cleaning up...")

	close(consumer.Props().Done)  // shutdown channel
	consumer.Props().Conn.Close() // close connection

	log.Println("rabbitMQ disconnection complete")
	return nil
}
