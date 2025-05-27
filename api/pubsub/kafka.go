package pubsub

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"

	"github.com/lzf-12/go-example-collections/msgbroker/adapter/kafka"
)

func InitKafkaConsumer(ctx context.Context, autoCreateTopic bool) {

	// TODO need to change to env, temporary hardcode
	kafkaBrokerHost := "localhost:9092"

	// admin configuration map
	admincfg := kafka.NewKafkaConfigMap()
	admincfg.Set(fmt.Sprintf("bootstrap.servers=%s", kafkaBrokerHost))

	consumerGroupId := "consumer-group-1"

	// consumer configuration map
	consumercfg := kafka.NewKafkaConfigMap()
	consumercfg.Set(fmt.Sprintf("bootstrap.servers=%s", kafkaBrokerHost))
	consumercfg.Set(fmt.Sprintf("group.id=%s", consumerGroupId))
	consumercfg.Set(fmt.Sprintf("auto.offset.reset=%s", "earliest"))

	log.Println("initialize kafka consumer client")
	kc, err := kafka.NewKafkaConsumerClient(consumercfg, admincfg)
	if err != nil {
		log.Println("error initialize kafka consumer client: ", err)
	}
	log.Println("success initialize kafka consumer client")

	// healtcheck
	log.Println("kafka healthcheck...")
	err = kc.HealthCheck(ctx)
	if err != nil {
		log.Println("healtcheck kafka error: ", err)
		return
	}
	log.Println("kafka ok")

	// TODO: centralize topic partition configuration in .yml
	// topic handlers map
	topicHandlers := []kafka.TopicHandler{
		{Topic: TopicOrderV2Json, Handler: orderHandlerV2Json, Partitions: 1, ReplicationFactor: 1},
		{Topic: TopicOrderV2Xml, Handler: orderHandlerV2Xml, Partitions: 1, ReplicationFactor: 1},
	}

	if autoCreateTopic {

		// create topics if not exist
		err = kc.CreateTopicsIfNotExist(ctx, topicHandlers)
		if err != nil {
			log.Println("error: ", err)
			return
		}

	}

	// subscribe all topic and handlers
	err = kc.SubscribeTopics(ctx, topicHandlers)
	if err != nil {
		log.Println("subscribe single topic error: ", err)
		return
	}
}

func orderHandlerV2Json(msg kafka.Message) error {
	var order OrderCreatedV2
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		return fmt.Errorf("failed to decode json order: %w", err)
	}

	// validate required fields
	if order.ID == "" || order.Product == "" || order.Quantity <= 0 {
		return fmt.Errorf("invalid order: missing required fields")
	}

	// business logic (e.g., save to DB, process payment, etc.)
	log.Printf("processing order: %+v", order)

	go func() {
		// process logic here

	}()

	return nil
}

func orderHandlerV2Xml(msg kafka.Message) error {
	var order OrderCreatedV2
	if err := xml.Unmarshal(msg.Value, &order); err != nil {
		return fmt.Errorf("failed to decode json order: %w", err)
	}

	// validate required fields
	if order.ID == "" || order.Product == "" || order.Quantity <= 0 {
		return fmt.Errorf("invalid order: missing required fields")
	}

	// business logic (e.g., save to DB, process payment, etc.)
	log.Printf("processing order: %+v", order)

	go func() {
		// process logic here

	}()

	return nil
}
