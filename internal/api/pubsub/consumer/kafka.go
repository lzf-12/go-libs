package consumer

import (
	"context"
	"fmt"
	"log"

	handler "github.com/lzf-12/go-example-collections/internal/api/pubsub/handler"
	pubsub "github.com/lzf-12/go-example-collections/internal/api/pubsub/model"
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
		{Topic: pubsub.TopicOrderV2Json, Handler: handler.OrderHandlerV2Json, Partitions: 1, ReplicationFactor: 1},
		{Topic: pubsub.TopicOrderV2Xml, Handler: handler.OrderHandlerV2Xml, Partitions: 1, ReplicationFactor: 1},
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
