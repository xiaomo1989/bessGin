package main

import (
	"bessGin/util/rabbitmq/fanout"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 初始化连接
	fanout.GetConnection()
	// 启动消费者
	go fanout.StartConsumer("CONSUMER-1")
	// 管理动态扩展
	//go autoScaleConsumers()
	// 优雅关闭
	handleShutdown()
	// 保持主协程运行
	select {}
}

func autoScaleConsumers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	consumerCount := 2
	for range ticker.C {
		consumerCount++
		log.Printf("Scaling out to %d consumers...", consumerCount)
		go fanout.StartConsumer(
			fmt.Sprintf("CONSUMER-%d", consumerCount),
		)
	}
}

func handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("\nStarting graceful shutdown...")
	if conn := fanout.GetConnection(); conn != nil {
		if err := conn.Close(); err != nil {
			log.Println("RabbitMQ close error:", err)
		}
	}
	log.Println("Resources released. Exiting.")
	os.Exit(0)
}
