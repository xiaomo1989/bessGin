package direct

import (
	"fmt"
	"os"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	once          sync.Once
	rabbitConn    *amqp.Connection
	rabbitChannel *amqp.Channel
)

const (
	ExchangeName = "app_events_direct"
	ExchangeType = "fanout"
)

func GetConnection() *amqp.Connection {
	once.Do(func() {
		var err error
		rabbitConn, err = amqp.Dial(getRabbitMQURL())
		if err != nil {
			panic(fmt.Sprintf("Failed to connect to RabbitMQ: %v", err))
		}

		rabbitChannel, err = rabbitConn.Channel()
		if err != nil {
			panic(fmt.Sprintf("Failed to open channel: %v", err))
		}

		err = rabbitChannel.ExchangeDeclare(
			ExchangeName,
			ExchangeType,
			true,  // durable
			false, // auto-delete
			false, // internal
			false, // no-wait
			nil,
		)
		if err != nil {
			panic(fmt.Sprintf("Failed to declare exchange: %v", err))
		}
	})
	return rabbitConn
}

func getRabbitMQURL() string {
	if url := os.Getenv("RABBITMQ_URL"); url != "" {
		return url
	}
	return "amqp://guest:guest@localhost:5672/"
}
