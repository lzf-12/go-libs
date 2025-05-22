package api

type OrderCreated struct {
	ID     string  `json:"id" xml:"id" avro:"id" msgpack:"id"`
	Amount float64 `json:"amount" xml:"amount" avro:"amount" msgpack:"amount"`
}

type QueueTopicHandler struct {
	Queue   string
	Topic   string
	Handler func([]byte, map[string]interface{})

	DeadLetterQueue string
}
