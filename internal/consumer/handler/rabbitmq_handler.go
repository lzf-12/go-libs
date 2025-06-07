package handler

import (
	"encoding/json"
	"encoding/xml"
	"log"

	"github.com/lzf-12/go-example-collections/internal/consumer/model"
)

func HandleCreateOrderV1JSON(msg []byte, _ map[string]interface{}) {
	var o model.OrderCreatedV1
	if err := json.Unmarshal(msg, &o); err != nil {
		log.Printf("[JSON] Failed to parse: %v", err)
		return
	}

	// call create order flow process here
}

func HandleCreateOrderV1XML(msg []byte, _ map[string]interface{}) {
	var o model.OrderCreatedV1
	if err := xml.Unmarshal(msg, &o); err != nil {
		log.Printf("[XML] Failed to parse: %v", err)
		return
	}

	// call create order flow process here
}
