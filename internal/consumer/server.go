package consumer

import (
	"context"
)

func ServeRabbitMQConsumer(ctx context.Context) error {

	if err := InitRabbitMQConsumer(ctx); err != nil {
		return err
	}

	return nil
}

func ServeKafkaConsumer(ctx context.Context) error {

	if err := InitKafkaConsumer(ctx); err != nil {
		return err
	}

	return nil
}
