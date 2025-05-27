package pubsub

import (
	"context"
)

func ServeRabbitMQConsumer(ctx context.Context) error {

	InitRabbitMQConsumer(ctx)

	return nil
}

func ServeKafkaConsumer(ctx context.Context) error {

	autoCreateTopicIfNotExist := true // TODO change to env variable
	InitKafkaConsumer(ctx, autoCreateTopicIfNotExist)

	return nil
}
