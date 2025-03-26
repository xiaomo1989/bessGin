package main

import (
	"github.com/streadway/amqp"
	"log"
	"time"
)

const (
	RabbitMQHhost    = "amqp://guest:guest@localhost:5672/"
	QueueNameHost    = "gin_queue"
	ConsumerNameHost = "gin_consumer"
)

func main() {
	// 每个消费者创建自己的连接和通道
	consumerID := "Consumer-3"
	conn, err := amqp.Dial(RabbitMQHhost)
	failError1(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failError1(err, "Failed to open a channel")
	defer ch.Close()

	// 声明队列
	q, err := ch.QueueDeclare(
		QueueNameHost,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failError1(err, "Failed to declare a queue")

	// 设置QoS（公平分发）
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failError1(err, "Failed to set QoS")

	// 注册消费者
	msgs, err := ch.Consume(
		q.Name,     // queue
		consumerID, // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	failError1(err, "Failed to register a consumer")

	log.Printf("Consumer %s started...", consumerID)

	// 处理消息
	for d := range msgs {
		log.Printf("%s received: %s", consumerID, d.Body)

		// 模拟处理耗时
		time.Sleep(1 * time.Second)

		// 手动确认消息
		d.Ack(false)
	}
}
func failError1(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
