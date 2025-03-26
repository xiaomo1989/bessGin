package fanout

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func Publish(message []byte) error {
	// 初始化连接
	GetConnection()
	/*conn := GetConnection()
	ch, _ := conn.Channel()
	*/
	return rabbitChannel.Publish(
		ExchangeName,
		"",    // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         message,
		},
	)
}
