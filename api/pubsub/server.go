package pubsub

import (
	"context"
)

func ServeRabbitMQConsumer(ctx context.Context) error {

	InitRabbitMQConsumer(ctx)

	return nil
}
