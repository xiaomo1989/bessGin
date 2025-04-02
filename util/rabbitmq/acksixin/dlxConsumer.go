package acksixin

import (
	"bessGin/config"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// DLXConsumer 死信队列消费者
type DLXConsumer struct {
	config      *config.RabbitMQConfig
	connManager *ConnectionManager
	channel     *amqp.Channel
	done        chan struct{}
	mu          sync.Mutex
	wg          sync.WaitGroup
}

// RabbitMQConfig 配置结构
type RabbitMQConfig struct {
	URL            string
	DLXQueueName   string
	DLXExchange    string
	PrefetchCount  int
	MaxRetries     int
	ReconnectDelay time.Duration
}

// NewDLXConsumer 创建新的消费者实例
func NewDLXConsumer(config *config.RabbitMQConfig) *DLXConsumer {
	return &DLXConsumer{
		config:      config,
		connManager: NewConnectionManager(config),
		done:        make(chan struct{}),
	}
}

// Start 启动消费者（带自动重连）
func (c *DLXConsumer) Start(ctx context.Context, handler func(msg amqp.Delivery) error) error {
	c.wg.Add(1)
	defer c.wg.Done()

	var retries int

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-c.done:
			return nil
		default:
			// 建立连接和通道
			if err := c.connect(); err != nil {
				if retries >= c.config.MaxRetries {
					return fmt.Errorf("max retries (%d) reached: %w", c.config.MaxRetries, err)
				}

				log.Printf("Connection failed, retrying in %v (attempt %d/%d)",
					c.config.ReconnectDelay, retries+1, c.config.MaxRetries)

				select {
				case <-time.After(c.config.ReconnectDelay):
				case <-ctx.Done():
					return nil
				}

				retries++
				continue
			}

			retries = 0 // 重置重试计数器

			// 开始消费
			if err := c.consumeDLX(handler); err != nil {
				log.Printf("Consuming error: %v", err)
				c.closeResources()
				continue
			}

			// 监听连接关闭事件
			connCloseChan := c.connManager.NotifyClose()
			channelCloseChan := c.channel.NotifyClose(make(chan *amqp.Error))

			select {
			case <-c.done:
				return nil
			case <-ctx.Done():
				return nil
			case err := <-connCloseChan:
				log.Printf("Connection closed: %v", err)
				c.closeResources()
			case err := <-channelCloseChan:
				log.Printf("Channel closed: %v", err)
				c.closeResources()
			}
		}
	}
}

// connect 建立连接和通道
func (c *DLXConsumer) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 获取连接
	conn, err := c.connManager.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	// 创建通道
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// 声明基础设施
	if err := declareDLXInfrastructure(ch, c.config); err != nil {
		_ = ch.Close()
		return fmt.Errorf("failed to declare infrastructure: %w", err)
	}

	// 设置QoS
	if err := ch.Qos(
		c.config.PrefetchCount,
		0,     // prefetchSize
		false, // global
	); err != nil {
		_ = ch.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	c.channel = ch
	return nil
}

// consumeDLX 开始消费死信队列
func (c *DLXConsumer) consumeDLX(handler func(msg amqp.Delivery) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel == nil {
		return errors.New("channel not initialized")
	}

	msgs, err := c.channel.Consume(
		c.config.DLXQueueName,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume: %w", err)
	}

	// 启动消息处理协程
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-c.done:
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("Message channel closed")
					return
				}

				if err := handler(msg); err != nil {
					log.Printf("Message handling failed: %v", err)
					_ = msg.Nack(false, true) // 重新入队
				} else {
					_ = msg.Ack(false)
				}
			}
		}
	}()

	log.Println("Started DLX consumer")
	return nil
}

// Shutdown 优雅关闭
func (c *DLXConsumer) Shutdown() {
	close(c.done)
	c.closeResources()
	c.wg.Wait()
	log.Println("DLX consumer shutdown complete")
}

// closeResources 关闭资源
func (c *DLXConsumer) closeResources() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil {
		_ = c.channel.Close()
		c.channel = nil
	}
}

// declareDLXInfrastructure 声明死信队列基础设施
func declareDLXInfrastructure(ch *amqp.Channel, cfg *config.RabbitMQConfig) error {
	// 声明DLX交换器
	if err := ch.ExchangeDeclare(
		cfg.DLXExchange,
		"fanout", // 死信通常用fanout广播
		true,     // durable
		false,    // autoDelete
		false,    // internal
		false,    // noWait
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare DLX exchange: %w", err)
	}

	// 声明DLX队列
	if _, err := ch.QueueDeclare(
		cfg.DLXQueueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		amqp.Table{
			"x-queue-type": "classic", // 或 "quorum" 用于高可用
		},
	); err != nil {
		return fmt.Errorf("failed to declare DLX queue: %w", err)
	}

	// 绑定DLX队列到交换器
	if err := ch.QueueBind(
		cfg.DLXQueueName,
		"", // fanout不需要routing key
		cfg.DLXExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind DLX queue: %w", err)
	}

	return nil
}
