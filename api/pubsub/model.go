package pubsub

import "time"

const (
	RmqQueueOrder    = "order-service.queue"
	TopicOrderV1Json = "order.v1.json"
	TopicOrderV1Xml  = "order.v1.xml"
	TopicOrderV2Json = "order.v2.json"
	TopicOrderV2Xml  = "order.v2.xml"
)

type OrderCreatedV1 struct {
	ID        string    `json:"id" xml:"id"`
	Product   string    `json:"product" xml:"product"`
	Quantity  int       `json:"quantity" xml:"quantity"`
	Price     float64   `json:"price" xml:"price"`
	Timestamp time.Time `json:"timestamp" xml:"timestamp"`
}

type OrderCreatedV2 struct {
	ID         string    `json:"id" xml:"id"`
	Product    string    `json:"product" xml:"product"`
	Quantity   int       `json:"quantity" xml:"quantity"`
	Price      float64   `json:"price" xml:"price"`
	Timestamp  time.Time `json:"timestamp" xml:"timestamp"`
	ConsumerId string    `json:"consumer_id" xml:"consumer_id"`
}

type QueueTopicHandler struct {
	Queue   string
	Topic   string
	Handler func([]byte, map[string]interface{})

	DeadLetterQueue string
}
