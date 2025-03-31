package fanout

import (
	"fmt"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	MainExchange = "app_main_exchange"
	LogQueue     = "message_log_queue"
	ExchangeType = "fanout"
)

var (
	connOnce   sync.Once
	rabbitConn *amqp.Connection
	messageSeq int64
	seqMutex   sync.Mutex
)

type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func GetConnection() *amqp.Connection {
	connOnce.Do(func() {
		var err error
		rabbitConn, err = amqp.Dial(getRabbitMQURL())
		if err != nil {
			panic(fmt.Sprintf("Failed to connect to RabbitMQ: %v", err))
		}
		initializeInfrastructure()
	})
	return rabbitConn
}

func initializeInfrastructure() {
	ch, err := rabbitConn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	// 主交换器（广播实时消息）
	err = ch.ExchangeDeclare(
		MainExchange,
		ExchangeType,
		true,  // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,
	)
	if err != nil {
		panic(err)
	}

	// 日志队列（持久化所有历史消息）
	_, err = ch.QueueDeclare(
		LogQueue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		panic(err)
	}
}

func getRabbitMQURL() string {
	if url := os.Getenv("RABBITMQ_URL"); url != "" {
		return url
	}
	return "amqp://guest:guest@localhost:5672/"
}

func GenerateMessageID() string {
	seqMutex.Lock()
	defer seqMutex.Unlock()
	messageSeq++
	return fmt.Sprintf("MSG-%d-%d", time.Now().UnixNano(), messageSeq)
}
