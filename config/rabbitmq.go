package config

import (
	"os"
	"strconv"
	"time"
)

type RabbitMQConfig struct {
	URL            string        `json:"url"`
	MainExchange   string        `json:"main_exchange"`
	RoutingKey     string        `json:"routing_key"`
	QueueName      string        `json:"queue_name"`
	DLXExchange    string        `json:"dlx_exchange"`
	DLXQueueName   string        `json:"dlx_queue_name"`
	PrefetchCount  int           `json:"prefetch_count"`
	ReconnectDelay time.Duration `json:"reconnect_delay"`
	MaxRetries     int           `json:"max_retries"`
	PublishTimeout time.Duration `json:"publish_timeout"`
	Heartbeat      time.Duration `json:"heartbeat"`
	ConnectTimeout time.Duration `json:"connect_timeout"`
}

func Load() *RabbitMQConfig {
	return &RabbitMQConfig{
		URL:            getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		MainExchange:   getEnv("RABBITMQ_MAIN_EXCHANGE", "orders.main"),
		RoutingKey:     getEnv("RABBITMQ_ROUTING_KEY", "order.created"),
		QueueName:      getEnv("RABBITMQ_QUEUE_NAME", "orders.queue"),
		DLXExchange:    getEnv("RABBITMQ_DLX_EXCHANGE", "orders.dlx"),
		DLXQueueName:   getEnv("RABBITMQ_DLX_QUEUE", "orders.dlx.queue"),
		PrefetchCount:  getEnvInt("RABBITMQ_PREFETCH_COUNT", 5),
		ReconnectDelay: getEnvDuration("RABBITMQ_RECONNECT_DELAY", 5*time.Second),
		MaxRetries:     getEnvInt("RABBITMQ_MAX_RETRIES", 3),
		PublishTimeout: getEnvDuration("RABBITMQ_PUBLISH_TIMEOUT", 10*time.Second),
		Heartbeat:      getEnvDuration("RABBITMQ_PUBLISH_TIMEOUT", 10*time.Second),
		ConnectTimeout: getEnvDuration("RABBITMQ_PUBLISH_TIMEOUT", 10*time.Second),
	}
}

// 辅助函数
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}
