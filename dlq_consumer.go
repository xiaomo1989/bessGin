// dlq_consumer.go
package main

import (
	"bessGin/config"
	"bessGin/util/rabbitmq/acksixin"
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

func main() {
	/*config := &RabbitMQConfig{
		URL:            "amqp://guest:guest@localhost:5672/",
		DLXQueueName:   "dlx.orders",
		DLXExchange:    "dlx_exchange",
		PrefetchCount:  10,
		MaxRetries:     5,
		ReconnectDelay: 3 * time.Second,
	}
	*/
	config := config.Load()
	consumer := acksixin.NewDLXConsumer(config)

	// 消息处理函数
	handler := func(msg amqp.Delivery) error {
		log.Printf("Received dead letter: %s", msg.Body)
		// 这里添加业务逻辑
		return nil
	}

	// 启动消费者
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := consumer.Start(ctx, handler); err != nil {
			log.Fatalf("Consumer failed: %v", err)
		}
	}()

	// 模拟运行一段时间后关闭
	time.Sleep(30 * time.Second)
	cancel()
	consumer.Shutdown()

}
