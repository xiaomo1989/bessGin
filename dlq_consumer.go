// dlq_consumer.go
package main

import (
	"bessGin/app/models"
	"bessGin/config"
	"bessGin/util/rabbitmq/acksixin"
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func main() {
	cfg := config.Load()
	ensureQueueExists(cfg) // 新增保障逻辑
	dlqCfg := &config.RabbitMQConfig{
		URL:           cfg.URL,
		QueueName:     cfg.DLXQueueName, // 直接消费死信队列
		PrefetchCount: 10,
	}

	// main.go
	consumer, err := acksixin.NewConsumer(dlqCfg, handleDLQMessage)
	if err != nil {
		log.Printf("DLQ Consumer创建失败详情: %+v", err) // 打印完整错误
		log.Fatalf("配置参数: %+v", dlqCfg)              // 输出当前配置
	}

	//consumer, err := acksixin.NewConsumer(dlqCfg, handleDLQMessage)
	//if err != nil {
	//	log.Fatal("DLQ consumer init failed:", err)
	//}

	consumer.Start(context.Background())
}

func handleDLQMessage(order models.Order) error {
	log.Printf("[DLQ] 收到死信消息 ID:%s", order.ID)

	// 实现以下逻辑：
	// 1. 持久化到数据库（如MongoDB）
	// 2. 发送报警（如企业微信）
	// 3. 记录完整上下文

	return nil // 消费后自动ACK
}

// main.go
func ensureQueueExists(cfg *config.RabbitMQConfig) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		log.Fatal("初始化连接失败:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("创建通道失败:", err)
	}
	defer ch.Close()

	// 声明死信队列
	_, err = ch.QueueDeclare(
		cfg.QueueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		amqp.Table{
			"x-dead-letter-exchange": cfg.DLXExchange, // 重要！
			"x-message-ttl":          60000,           // 60秒过期
		},
	)
	if err != nil {
		log.Fatalf("队列声明失败: %v", err)
	}
}
