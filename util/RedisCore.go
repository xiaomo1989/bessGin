package util

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	_ "github.com/redis/go-redis/v9"
	"log"
	"sync"
	"time"
)

var (
	rdb       *redis.Client
	onceRedis sync.Once
	ctx       = context.Background()
)

// RedisClient 结构体（封装 Redis 连接和方法）
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// GetRedis 返回 Redis 单例实例
func GetRedis() *RedisClient {
	onceRedis.Do(func() {
		rdb = redis.NewClient(&redis.Options{
			Addr:     "127.0.0.1:6379", // Redis 地址
			Password: "",               // 如果有密码，填入这里
			DB:       0,                // 使用的数据库索引
		})

		// 测试连接
		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			log.Fatalf("Redis 连接失败: %v", err)
		}
		fmt.Println("Redis 连接成功！")
	})
	return &RedisClient{client: rdb, ctx: ctx}
}

// Set 设置值
func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// Get 获取值
func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

// Del 删除键
func (r *RedisClient) Del(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// Exists 判断键是否存在
func (r *RedisClient) Exists(key string) (bool, error) {
	exists, err := r.client.Exists(r.ctx, key).Result()
	return exists > 0, err
}

// Incr 自增
func (r *RedisClient) Incr(key string) (int64, error) {
	return r.client.Incr(r.ctx, key).Result()
}

// HSet 设置哈希值
func (r *RedisClient) HSet(key string, field string, value interface{}) error {
	return r.client.HSet(r.ctx, key, field, value).Err()
}

// HGet 获取哈希值
func (r *RedisClient) HGet(key string, field string) (string, error) {
	return r.client.HGet(r.ctx, key, field).Result()
}

// Publish 发布消息
func (r *RedisClient) Publish(channel string, message string) error {
	return r.client.Publish(r.ctx, channel, message).Err()
}

// Subscribe 方法：订阅指定频道，并返回接收到的消息
func (r *RedisClient) SubscribeReturn(channel string) <-chan string {
	// 创建 Redis 订阅
	sub := rdb.Subscribe(ctx, channel)
	// 获取消息通道
	ch := make(chan string)
	// 在协程中监听消息
	go func() {
		defer close(ch)
		defer sub.Close()
		// 监听 Redis 频道消息
		for msg := range sub.Channel() {
			log.Printf("收到 Redis 消息: %s", msg.Payload)
			ch <- msg.Payload // 将消息传递到外部
		}
	}()

	return ch
}

// Subscribe 订阅消息
func (r *RedisClient) Subscribe(channel string, handler func(msg string)) {
	pubsub := r.client.Subscribe(r.ctx, channel)
	ch := pubsub.Channel()

	// 监听消息
	for msg := range ch {
		handler(msg.Payload)
	}
}
