package util

import (
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

const (
	RabbitMQURL  = "amqp://guest:guest@192.168.35.13:5672/" // RabbitMQ 连接地址
	ExchangeName = "exchange_demo_yan"                      // 交换机名称
	QueueName    = "queue_demo_yan"                         // 队列名称
	RoutingKey   = "routing_demo"                           // 路由键
)

var (
	instance *RabbitMQ
	once     sync.Once
)

// RabbitMQ 结构体
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// 获取 RabbitMQ 单例
func GetRabbitMQInstance() *RabbitMQ {
	once.Do(func() {
		instance = &RabbitMQ{}
		err := instance.connect()
		if err != nil {
			log.Fatalf("RabbitMQ 连接失败: %v", err)
		}
	})
	return instance
}

func (r *RabbitMQ) handleReconnect() {
	closeChan := r.conn.NotifyClose(make(chan *amqp.Error))
	go func() {
		err := <-closeChan
		if err != nil {
			log.Println("RabbitMQ 连接断开，尝试重新连接...")
			for {
				err := r.connect()
				if err == nil {
					log.Println("RabbitMQ 重新连接成功")
					break
				}
				log.Printf("RabbitMQ 重新连接失败: %v", err)
				time.Sleep(3 * time.Second) // 3秒后重试
			}
		}
	}()
}

// 连接 RabbitMQ
func (r *RabbitMQ) connect() error {
	var err error
	r.conn, err = amqp.Dial(RabbitMQURL)
	if err != nil {
		return err
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		return err
	}

	// 监听连接状态，自动重连
	r.handleReconnect()

	// 声明交换机、队列等
	err = r.channel.ExchangeDeclare(ExchangeName, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	_, err = r.channel.QueueDeclare(QueueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = r.channel.QueueBind(QueueName, "", ExchangeName, false, nil)
	if err != nil {
		return err
	}

	log.Println("RabbitMQ 连接成功")
	return nil
}

// 检查 RabbitMQ 连接是否可用，如果不可用则重连
func (r *RabbitMQ) ensureConnection() error {
	if r.conn == nil || r.conn.IsClosed() {
		log.Println("RabbitMQ 连接已关闭，尝试重新连接...")
		return r.connect()
	}

	// 监听通道关闭
	errChan := make(chan *amqp.Error, 1)
	r.channel.NotifyClose(errChan)
	select {
	case err := <-errChan:
		log.Printf("RabbitMQ 通道关闭: %v，正在尝试重连...", err)
		return r.connect()
	default:
	}
	return nil
}

// 发布消息
func (r *RabbitMQ) Publish(message string) error {
	if err := r.ensureConnection(); err != nil {
		return err
	}
	return r.channel.Publish(
		ExchangeName, "", false, false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
}

func (r *RabbitMQ) Subscribe() (<-chan amqp.Delivery, error) {
	q, err := r.channel.QueueDeclare(
		QueueName, // 使用固定队列名称
		true,      // durable=true（持久化）
		false,     // autoDelete=false（不自动删除）
		false,     // exclusive=false（非独占队列，允许多个消费者）
		false,     // noWait=false
		nil,
	)
	if err != nil {
		return nil, err
	}

	// 绑定队列到 fanout 交换机
	err = r.channel.QueueBind(
		q.Name, "", ExchangeName, false, nil,
	)
	if err != nil {
		return nil, err
	}

	// 订阅队列消息
	msgs, err := r.channel.Consume(
		q.Name, "", true, false, false, false, nil,
	)
	if err != nil {
		return nil, err
	}

	log.Printf("成功订阅消息，队列: %s", q.Name)
	return msgs, nil
}

// 关闭 RabbitMQ 连接
func (r *RabbitMQ) Close() {
	r.channel.Close()
	r.conn.Close()
}
