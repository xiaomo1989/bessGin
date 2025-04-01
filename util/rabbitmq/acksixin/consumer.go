package acksixin

import (
	"bessGin/app/models"
	"bessGin/config"
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type Consumer struct {
	cfg          *config.RabbitMQConfig
	connMgr      *ConnectionManager
	channel      *amqp.Channel
	deliveryChan <-chan amqp.Delivery
	handler      func(models.Order) error
}

func NewConsumer(
	cfg *config.RabbitMQConfig,
	handler func(models.Order) error,
) (*Consumer, error) {
	connMgr := NewConnectionManager(cfg)
	conn, err := connMgr.GetConnection()
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	if err := ch.Qos(cfg.PrefetchCount, 0, false); err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	deliveries, err := ch.Consume(
		cfg.QueueName,
		"",
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start consuming: %w", err)
	}

	return &Consumer{
		cfg:          cfg,
		connMgr:      connMgr,
		channel:      ch,
		deliveryChan: deliveries,
		handler:      handler,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	log.Println("Starting consumer...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down consumer...")
			return
		case d, ok := <-c.deliveryChan:
			if !ok {
				log.Println("Delivery channel closed, reconnecting...")
				if err := c.reconnect(); err != nil {
					log.Printf("Reconnect failed: %v", err)
				}
				continue
			}
			c.processDelivery(d)
		}
	}
}

func (c *Consumer) processDelivery(d amqp.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			c.nack(d, true)
		}
	}()

	var order models.Order
	if err := json.Unmarshal(d.Body, &order); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		c.nack(d, false)
		return
	}

	attempt := getAttemptCount(d.Headers)
	if attempt > c.cfg.MaxRetries {
		log.Printf("Message %s exceeded max retries (%d)", order.ID, c.cfg.MaxRetries)
		c.nack(d, false)
		return
	}

	if err := c.handler(order); err != nil {
		log.Printf("Processing failed (attempt %d): %v", attempt, err)
		c.republish(d)
		return
	}

	if err := d.Ack(false); err != nil {
		log.Printf("Failed to ack message: %v", err)
	}
}

func (c *Consumer) republish(d amqp.Delivery) {
	headers := amqp.Table{}
	for k, v := range d.Headers {
		headers[k] = v
	}
	headers["x-attempt"] = getAttemptCount(d.Headers) + 1

	err := c.channel.Publish(
		c.cfg.MainExchange,
		d.RoutingKey,
		true,
		false,
		amqp.Publishing{
			ContentType:  d.ContentType,
			Body:         d.Body,
			DeliveryMode: d.DeliveryMode,
			Headers:      headers,
		},
	)

	if err != nil {
		log.Printf("Failed to republish: %v", err)
		c.nack(d, true)
	} else {
		c.ack(d)
	}
}

func (c *Consumer) nack(d amqp.Delivery, requeue bool) {
	if err := d.Nack(false, requeue); err != nil {
		log.Printf("Failed to nack message: %v", err)
	}
}

func (c *Consumer) ack(d amqp.Delivery) {
	if err := d.Ack(false); err != nil {
		log.Printf("Failed to ack message: %v", err)
	}
}

func (c *Consumer) reconnect() error {
	conn, err := c.connMgr.GetConnection()
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	deliveries, err := ch.Consume(
		c.cfg.QueueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	c.channel = ch
	c.deliveryChan = deliveries
	return nil
}

func getAttemptCount(headers amqp.Table) int {
	if attempt, ok := headers["x-attempt"].(int32); ok {
		return int(attempt)
	}
	return 0
}

func (c *Consumer) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return fmt.Errorf("channel close failed: %w", err)
		}
	}
	return nil
}
