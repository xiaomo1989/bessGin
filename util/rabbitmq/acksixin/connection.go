package acksixin

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"bessGin/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	conn      *amqp.Connection
	connMutex sync.RWMutex
)

type ConnectionManager struct {
	config     *config.RabbitMQConfig
	connNotify chan *amqp.Error
	conn       *amqp.Connection
	mu         sync.RWMutex
	closed     bool
}

// NotifyClose 返回连接关闭通知通道
func (cm *ConnectionManager) NotifyClose() <-chan *amqp.Error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.conn != nil {
		return cm.conn.NotifyClose(make(chan *amqp.Error, 1))
	}

	ch := make(chan *amqp.Error, 1)
	close(ch)
	return ch
}

func NewConnectionManager(cfg *config.RabbitMQConfig) *ConnectionManager {
	return &ConnectionManager{
		config:     cfg,
		connNotify: make(chan *amqp.Error, 1),
	}
}

/*func (cm *ConnectionManager) GetConnection() (*amqp.Connection, error) {
	connMutex.RLock()
	if conn != nil && !conn.IsClosed() {
		connMutex.RUnlock()
		return conn, nil
	}
	connMutex.RUnlock()

	return cm.connectWithRetry()
}*/

// 在 ConnectionManager 中添加重试机制
func (cm *ConnectionManager) GetConnection() (*amqp.Connection, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.closed {
		return nil, errors.New("connection manager is closed")
	}

	// 如果连接有效则直接返回
	if cm.conn != nil && !cm.conn.IsClosed() {
		return cm.conn, nil
	}

	// 带指数退避的重试逻辑
	var lastErr error
	for i := 0; i < cm.config.MaxRetries; i++ {
		conn, err := amqp.DialConfig(cm.config.URL, amqp.Config{
			Heartbeat: cm.config.Heartbeat,
			Dial:      amqp.DefaultDial(cm.config.ConnectTimeout),
		})
		if err == nil {
			cm.conn = conn
			return conn, nil
		}
		lastErr = err
		time.Sleep(time.Second * time.Duration(math.Pow(2, float64(i))))
	}
	return nil, fmt.Errorf("after %d retries: %w", cm.config.MaxRetries, lastErr)
}

func (cm *ConnectionManager) connectWithRetry() (*amqp.Connection, error) {
	var lastErr error
	for i := 0; i < cm.config.MaxRetries; i++ {
		c, err := amqp.DialConfig(cm.config.URL, amqp.Config{
			Heartbeat: 10 * time.Second,
			Locale:    "en_US",
		})
		if err == nil {
			connMutex.Lock()
			conn = c
			connMutex.Unlock()

			cm.monitorConnection()
			return conn, nil
		}
		lastErr = err
		time.Sleep(cm.config.ReconnectDelay)
	}
	return nil, fmt.Errorf("failed to connect after %d attempts: %v", cm.config.MaxRetries, lastErr)
}

// 完整基础设施声明
func declareInfrastructure(ch *amqp.Channel, cfg *config.RabbitMQConfig) error {
	// 声明主交换器
	if err := ch.ExchangeDeclare(
		cfg.MainExchange,
		"direct",
		true,  // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare main exchange: %w", err)
	}

	// 声明死信交换器
	if err := ch.ExchangeDeclare(
		cfg.DLXExchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare DLX: %w", err)
	}

	// 声明主队列（带死信设置）
	_, err := ch.QueueDeclare(
		cfg.QueueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		amqp.Table{
			"x-dead-letter-exchange": cfg.DLXExchange,
			"x-message-ttl":          60000, // 60秒过期
			"x-max-retries":          cfg.MaxRetries,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare main queue: %w", err)
	}

	// 绑定主队列到交换器
	if err := ch.QueueBind(
		cfg.QueueName,
		cfg.RoutingKey,
		cfg.MainExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind main queue: %w", err)
	}

	// 声明死信队列
	_, err = ch.QueueDeclare(
		cfg.DLXQueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLX queue: %w", err)
	}

	// 绑定死信队列
	return ch.QueueBind(
		cfg.DLXQueueName,
		"", // fanout交换器不需要routing key
		cfg.DLXExchange,
		false,
		nil,
	)
}

func (cm *ConnectionManager) monitorConnection() {
	connMutex.RLock()
	defer connMutex.RUnlock()

	conn.NotifyClose(cm.connNotify)
	go func() {
		err := <-cm.connNotify
		log.Printf("Connection closed: %v, attempting reconnect...", err)
		_, _ = cm.connectWithRetry()
	}()
}
