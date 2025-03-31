package fanout

import (
	"bessGin/app/service"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
	"time"
)

type Consumer struct {
	ID           string
	processedIDs sync.Map
}

func NewConsumer(id string) *Consumer {
	return &Consumer{
		ID: id,
	}
}

func (c *Consumer) Start() {
	conn := GetConnection()
	// 第一步：重放历史消息
	c.replayHistoricalMessages(conn)
	// 第二步：监听实时消息
	go c.listenRealtimeMessages(conn)
}

func (c *Consumer) replayHistoricalMessages(conn *amqp.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("[%s] Channel error: %v", c.ID, err)
		return
	}
	defer ch.Close()

	// 获取日志队列信息
	q, err := ch.QueueInspect(LogQueue)
	if err != nil {
		log.Printf("[%s] Queue inspect error: %v", c.ID, err)
		return
	}

	log.Printf("[%s] Replaying %d historical messages...", c.ID, q.Messages)

	// 消费所有历史消息
	msgs, err := ch.Consume(
		LogQueue,
		c.ID+"_replay",
		false, // autoAck
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("[%s] Consume error: %v", c.ID, err)
		return
	}

	for d := range msgs {
		var msg Message
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			log.Printf("[%s] Decode error: %v", c.ID, err)
			continue
		}

		if _, exists := c.processedIDs.Load(msg.ID); !exists {
			c.processMessage(msg, "HISTORY")
			c.processedIDs.Store(msg.ID, struct{}{})
		}

		d.Ack(false)

		// 简单退出条件（实际应更严谨）
		if q.Messages == 0 {
			break
		}
	}
}

func (c *Consumer) listenRealtimeMessages(conn *amqp.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("[%s] Channel error: %v", c.ID, err)
		return
	}
	defer ch.Close()

	// 创建临时队列并绑定到主交换器
	q, err := ch.QueueDeclare(
		"",    // 随机队列名
		false, // durable
		true,  // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		log.Printf("[%s] Queue declare error: %v", c.ID, err)
		return
	}

	err = ch.QueueBind(
		q.Name,
		"", // routing key
		MainExchange,
		false,
		nil,
	)
	if err != nil {
		log.Printf("[%s] Bind error: %v", c.ID, err)
		return
	}

	msgs, err := ch.Consume(
		q.Name,
		c.ID+"_realtime",
		false, // autoAck
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("[%s] Consume error: %v", c.ID, err)
		return
	}

	log.Printf("[%s] Listening for real-time messages...", c.ID)

	for d := range msgs {
		var msg Message
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			log.Printf("[%s] Decode error: %v", c.ID, err)
			continue
		}

		if _, exists := c.processedIDs.Load(msg.ID); !exists {
			c.processMessage(msg, "REALTIME")
			c.processedIDs.Store(msg.ID, struct{}{})
		}

		d.Ack(false)
	}
}

func (c *Consumer) processMessage(msg Message, source string) {
	// 模拟业务处理
	log.Printf("[%s] Processed %s message: %s (ID: %s)",
		c.ID,
		source,
		msg.Content,
		msg.ID,
	)

	// 错误方式2: 值类型实例调用指针接收者的方法（可能报错）
	var msgService service.MessageService
	_ = msgService.SaveRecord() // Go 会自动取地址，但代码需符合预期

	// 添加实际业务逻辑
	time.Sleep(100 * time.Millisecond) // 模拟处理时间
}

func StartConsumer(id string) {
	c := NewConsumer(id)
	c.Start()
}
