package acksixin

import (
	"bessGin/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type DLXConsumer struct {
	config     *config.RabbitMQConfig
	connection *amqp.Connection
	channel    *amqp.Channel
	done       chan struct{}
}

func NewDLXConsumer(config *config.RabbitMQConfig) *DLXConsumer {
	return &DLXConsumer{
		config: config,
		done:   make(chan struct{}),
	}
}

// 启动消费者（带自动重连）
func (c *DLXConsumer) Start(ctx context.Context) error {
	var retries int

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := c.connect()
			if err != nil {
				if retries >= c.config.MaxRetries {
					return fmt.Errorf("max retries (%d) reached", c.config.MaxRetries)
				}

				log.Printf("Connection failed, retrying in %v (attempt %d/%d)",
					c.config.ReconnectDelay, retries+1, c.config.MaxRetries)

				time.Sleep(c.config.ReconnectDelay)
				retries++
				continue
			}

			retries = 0 // 重置重试计数器
			log.Println("Connected to RabbitMQ")

			err = c.consumeDLX()
			if err != nil {
				log.Printf("Consuming error: %v", err)
				c.close()
				continue
			}

			// 等待连接关闭或主动终止
			select {
			case <-c.done:
				return nil
			case <-c.connection.NotifyClose(make(chan *amqp.Error)):
				log.Println("Connection closed, reconnecting...")
			}
		}
	}
}

// 建立连接和通道
func (c *DLXConsumer) connect() error {
	var err error

	c.connection, err = amqp.DialConfig(c.config.URL, amqp.Config{
		Dial: amqp.DefaultDial(c.config.PublishTimeout),
	})
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	c.channel, err = c.connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// 声明DLX队列（确保存在）
	_, err = c.channel.QueueDeclare(
		c.config.DLXQueueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLX queue: %w", err)
	}

	// 设置QoS
	err = c.channel.Qos(
		c.config.PrefetchCount,
		0,     // prefetchSize
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	return nil
}

// 消费死信队列
func (c *DLXConsumer) consumeDLX() error {
	msgs, err := c.channel.Consume(
		c.config.DLXQueueName,
		"",    // consumer
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume DLX queue: %w", err)
	}

	log.Printf("Started consuming DLX queue: %s", c.config.DLXQueueName)

	for {
		select {
		case <-c.done:
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("message channel closed")
			}

			// 处理死信消息
			if err := c.handleMessage(msg); err != nil {
				log.Printf("Failed to handle message: %v", err)
				// 可以选择将消息转移到其他队列或记录日志
				_ = msg.Nack(false, false) // 不重试
			} else {
				_ = msg.Ack(false)
			}
		}
	}
}

// 处理消息（示例实现）
func (c *DLXConsumer) handleMessage(msg amqp.Delivery) error {
	// 这里实现你的业务逻辑
	log.Printf("Received dead letter message: %s", msg.Body)

	// 示例：检查重试次数
	headers := msg.Headers
	retryCount, _ := headers["x-retry-count"].(int32)
	if retryCount > int32(c.config.MaxRetries) {
		log.Printf("Message exceeded max retries: %d", retryCount)
		return fmt.Errorf("max retries exceeded")
	}

	// TODO: 添加你的处理逻辑

	return nil
}

// 关闭连接
func (c *DLXConsumer) close() {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.connection != nil {
		_ = c.connection.Close()
	}
	close(c.done)
}

// 优雅关闭
func (c *DLXConsumer) Shutdown() {
	c.close()
	log.Println("DLX consumer stopped")
}
