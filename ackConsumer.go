package main

import (
	"bessGin/app/models"
	"bessGin/config"
	"bessGin/util/rabbitmq/acksixin"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// 加载配置
	cfg := config.Load()

	// 初始化消费者
	consumer, err := acksixin.NewConsumer(cfg, handleOrder)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Consumer shutdown error: %v", err)
		}
	}()

	// 上下文管理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("\nReceived shutdown signal")
		cancel()
	}()

	//go dlqConsumer.Start(ctx)

	// 启动消费
	consumer.Start(ctx)
	log.Println("Consumer service running...")
	<-ctx.Done()
	log.Println("Consumer shutdown completed")
}

func handleOrder(order models.Order) error {
	// 模拟业务处理
	time.Sleep(time.Duration(300+rand.Intn(700)) * time.Millisecond)

	// 模拟随机失败（10%失败率）
	if rand.Float32() < 0.1 {
		return errors.New("random processing failure")
	}

	log.Printf("Successfully processed order %s", order.ID)
	return nil
}

func archiveToFile(order models.Order) error {
	f, err := os.OpenFile("dlq_archive.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	entry := fmt.Sprintf("[%s] %+v\n", time.Now().Format(time.RFC3339), order)
	if _, err := f.WriteString(entry); err != nil {
		return err
	}
	return nil
}
