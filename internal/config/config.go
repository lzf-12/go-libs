package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	defaultDLQ = "default.dlq"
	defaultDLX = "default.dlx"
)

type EnvVal struct {
	stringVal string
}

type Config struct {
	RabbitMQAmqpString string
	RabbitMQDefaultDlq string
	RabbitMQDefaultDlx string

	KafkaHost            string
	KafkaPort            string
	KafkaConsumerGroupID string
}

func LoadConfig(path string) (*Config, error) {

	if err := godotenv.Load(path); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	cfg := &Config{
		//rabbitmq
		RabbitMQAmqpString: getEnv("RABBITMQ_AMQP_STRING", "").String(),
		RabbitMQDefaultDlq: getEnv("RABBITMQ_DEFAULT_DLQ", defaultDLQ).String(),
		RabbitMQDefaultDlx: getEnv("RABBITMQ_DEFAULT_DLX", defaultDLX).String(),

		//kafka
		KafkaHost:            getEnv("KAFKA_HOST", "").String(),
		KafkaPort:            getEnv("KAFKA_PORT", "").String(),
		KafkaConsumerGroupID: getEnv("KAFKA_CONSUMER_GROUP_ID", "consumer-group-1").String(),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) EnvVal {
	if stringVal, exists := os.LookupEnv(key); exists {
		return EnvVal{stringVal: stringVal}
	}
	return EnvVal{stringVal: defaultValue}
}

func (ev EnvVal) String() string {
	return ev.stringVal
}

func (ev EnvVal) Bool() bool {
	val := strings.ToLower(ev.stringVal)
	return val == "true" || val == "1" || val == "yes" || val == "on"
}

func (ev EnvVal) Int() (int, error) {
	return strconv.Atoi(ev.stringVal)
}

func (ev EnvVal) IntDefault(defaultValue int) int {
	if val, err := strconv.Atoi(ev.stringVal); err == nil {
		return val
	}
	return defaultValue
}

func (ev EnvVal) Float64() (float64, error) {
	return strconv.ParseFloat(ev.stringVal, 64)
}

func (ev EnvVal) Float64Default(defaultValue float64) float64 {
	if val, err := strconv.ParseFloat(ev.stringVal, 64); err == nil {
		return val
	}
	return defaultValue
}

func (ev EnvVal) StringSlice(sep string) []string {
	if ev.stringVal == "" {
		return []string{}
	}
	return strings.Split(ev.stringVal, sep)
}
