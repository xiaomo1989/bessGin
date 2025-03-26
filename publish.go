package main

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

const (
	queueName    = "task_queue"
	amqpURL      = "amqp://guest:guest@localhost:5672/"
	maxAttempts  = 3 // 最大重试次数
	RabbitMQURL  = "amqp://guest:guest@localhost:5672/"
	QueueName    = "gin_queue"
	ConsumerName = "gin_consumer"
)

func main() {
	// 初始化RabbitMQ连接
	conn, err := amqp.Dial(RabbitMQURL)
	failError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failError(err, "Failed to open a channel")
	defer ch.Close()

	// 声明队列（持久化）
	_, err = ch.QueueDeclare(
		QueueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failError(err, "Failed to declare a queue")

	message := "我是消息"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 发布消息
	// 修正后的代码片段
	err = ch.PublishWithContext(ctx,
		"",        // exchange
		QueueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message),
		})

	failError(err, "Failed to publish a message")
}

func failError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
		fmt.Println(err)
	}
}
