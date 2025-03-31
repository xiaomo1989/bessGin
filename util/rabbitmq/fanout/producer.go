package fanout

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

func PublishMessage(content string) error {
	conn := GetConnection()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// 构造消息
	msg := Message{
		ID:        GenerateMessageID(),
		Content:   content,
		Timestamp: time.Now(),
	}

	// 序列化消息
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// 双写机制：同时写入主交换器和日志队列
	err = publishToMainExchange(ch, body)
	if err != nil {
		return err
	}
	err = publishToLogQueue(ch, body)
	if err != nil {
		return err
	}

	return nil
}

func publishToMainExchange(ch *amqp.Channel, body []byte) error {
	return ch.PublishWithContext(
		context.Background(),
		MainExchange,
		"", // routing key
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}

func publishToLogQueue(ch *amqp.Channel, body []byte) error {
	return ch.PublishWithContext(
		context.Background(),
		"",       // exchange
		LogQueue, // routing key
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}
