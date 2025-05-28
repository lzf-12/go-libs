package pubsub

import (
	"context"

	"github.com/lzf-12/go-example-collections/internal/api/pubsub/consumer"
)

func ServeRabbitMQConsumer(ctx context.Context) error {

	consumer.InitRabbitMQConsumer(ctx)

	return nil
}

func ServeKafkaConsumer(ctx context.Context) error {

	autoCreateTopicIfNotExist := true // TODO change to env variable
	consumer.InitKafkaConsumer(ctx, autoCreateTopicIfNotExist)

	return nil
}
