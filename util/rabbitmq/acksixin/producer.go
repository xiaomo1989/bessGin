package acksixin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bessGin/app/models"
	"bessGin/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	cfg       *config.RabbitMQConfig
	connMgr   *ConnectionManager
	channel   *amqp.Channel
	confirmCh chan amqp.Confirmation
}

func NewProducer(cfg *config.RabbitMQConfig) (*Producer, error) {
	connMgr := NewConnectionManager(cfg)
	conn, err := connMgr.GetConnection()
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		return nil, fmt.Errorf("failed to enable confirm mode: %w", err)
	}

	if err := declareInfrastructure(ch, cfg); err != nil {
		return nil, err
	}

	return &Producer{
		cfg:       cfg,
		connMgr:   connMgr,
		channel:   ch,
		confirmCh: ch.NotifyPublish(make(chan amqp.Confirmation, 1)),
	}, nil
}

func (p *Producer) PublishOrder(ctx context.Context, order models.Order) error {
	if err := p.ensureChannel(); err != nil {
		return err
	}

	body, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, p.cfg.PublishTimeout)
	defer cancel()

	return p.publishWithRetry(ctx, body)
}

func (p *Producer) publishWithRetry(ctx context.Context, body []byte) error {
	for attempt := 1; ; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := p.channel.PublishWithContext(
				ctx,
				p.cfg.MainExchange,
				p.cfg.RoutingKey,
				true,  // mandatory
				false, // immediate
				amqp.Publishing{
					DeliveryMode: amqp.Persistent,
					ContentType:  "application/json",
					Body:         body,
					Timestamp:    time.Now(),
					Headers: amqp.Table{
						"x-attempt": attempt,
					},
				},
			)

			if err != nil {
				if attempt > p.cfg.MaxRetries {
					return fmt.Errorf("max retries exceeded: %w", err)
				}
				if err = p.reconnect(); err != nil {
					return err
				}
				continue
			}

			confirmed, ok := <-p.confirmCh
			if !ok || !confirmed.Ack {
				return errors.New("broker failed to acknowledge message")
			}
			return nil
		}
	}
}

func (p *Producer) ensureChannel() error {
	if p.channel == nil || p.channel.IsClosed() {
		conn, err := p.connMgr.GetConnection()
		if err != nil {
			return err
		}

		ch, err := conn.Channel()
		if err != nil {
			return err
		}
		p.channel = ch
	}
	return nil
}

func (p *Producer) reconnect() error {
	if _, err := p.connMgr.connectWithRetry(); err != nil {
		return err
	}
	return p.ensureChannel()
}

func (p *Producer) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			return fmt.Errorf("channel close failed: %w", err)
		}
	}
	return nil
}
