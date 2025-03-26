package consumer

import (
	"log"
)

type Consumer struct {
	ID        string
	queueName string
}

func NewConsumer(id string) *Consumer {
	return &Consumer{
		ID: id,
	}
}

func (c *Consumer) Start() {
	conn := GetConnection()
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("%s: failed to open channel: %v", c.ID, err)
		return
	}
	defer ch.Close()

	// 创建临时队列
	q, err := ch.QueueDeclare(
		"",    // 随机生成队列名
		false, // durable
		true,  // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		log.Printf("%s: failed to declare queue: %v", c.ID, err)
		return
	}

	// 绑定到交换器
	err = ch.QueueBind(
		q.Name,
		"",           // routing key
		"app_events", // exchange name
		false,
		nil,
	)
	if err != nil {
		log.Printf("%s: failed to bind queue: %v", c.ID, err)
		return
	}

	// 开始消费
	msgs, err := ch.Consume(
		q.Name,
		c.ID,
		true,  // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		log.Printf("%s: failed to consume: %v", c.ID, err)
		return
	}

	log.Printf("%s started (Queue: %s)", c.ID, q.Name)

	for msg := range msgs {
		log.Printf("%s received: %s", c.ID, msg.Body)
	}
}
