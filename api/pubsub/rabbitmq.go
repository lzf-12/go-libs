package api

import (
	"encoding/json"
	"encoding/xml"
	"log"
	"time"

	"github.com/lzf-12/go-example-collections/msgbroker/adapter/rabbitmq"
	"github.com/lzf-12/go-example-collections/msgbroker/retry"
)

const (
	QueueOrder       = "order-service.queue"
	TopicOrderV1Json = "order.v1.json"
	TopicOrderV1Xml  = "order.v1.xml"
)

func InitConsumer() {

	opts := rabbitmq.RabbitMQOpts{
		AmqpString: "amqp://guest:guest@localhost:5672/",
	}

	rmq, err := rabbitmq.NewRabbitMQBroker(opts)
	if err != nil {
		log.Fatalf("rabbitMQ initialize connection failed: %v", err)
	}

	consumerCfg := rabbitmq.ConsumerCfg{
		PrefetchCount: 0,
		PrefetchSize:  1,
		RetryPolicy: retry.RetryPolicy{
			MaxRetries:      3,
			InitialInterval: 2 * time.Second,
			Multiplier:      2,
			MaxInterval:     10 * time.Second,
		},
		DLQExchange:   "default.dlx",
		DLQRoutingKey: "default.dlq",
	}

	consumer, err := rmq.NewConsumer(consumerCfg)
	if err != nil {
		log.Fatalf("consumer initialize failed: %v", err)
	}

	mapQueueTopicHandler := []QueueTopicHandler{
		{
			Queue:   QueueOrder,
			Topic:   TopicOrderV1Json,
			Handler: handleCreateOrderJSON,
		},
		{
			Queue:   QueueOrder,
			Topic:   TopicOrderV1Xml,
			Handler: handleCreateOrderXML,
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

}

func handleCreateOrderJSON(msg []byte, _ map[string]interface{}) {
	var o OrderCreated
	if err := json.Unmarshal(msg, &o); err != nil {
		log.Printf("[JSON] Failed to parse: %v", err)
		return
	}

	// call create order flow process here
}

func handleCreateOrderXML(msg []byte, _ map[string]interface{}) {
	var o OrderCreated
	if err := xml.Unmarshal(msg, &o); err != nil {
		log.Printf("[XML] Failed to parse: %v", err)
		return
	}

	// call create order flow process here
}
