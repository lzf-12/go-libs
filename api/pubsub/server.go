package api

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func ServeRabbitMQPubsub() {

	InitConsumer()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down consumer...")

}
