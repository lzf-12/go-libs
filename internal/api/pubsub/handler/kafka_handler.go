package handler

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"

	"github.com/lzf-12/go-example-collections/msgbroker/adapter/kafka"

	"github.com/lzf-12/go-example-collections/internal/api/pubsub/model"
)

func OrderHandlerV2Json(msg kafka.Message) error {
	var order model.OrderCreatedV2
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

func OrderHandlerV2Xml(msg kafka.Message) error {
	var order model.OrderCreatedV2
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
