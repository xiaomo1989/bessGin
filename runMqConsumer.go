package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bessGin/util/rabbitmq/fanout"
)

func main() {
	// 初始化连接
	fanout.GetConnection()

	// 启动初始消费者
	startConsumers(1)

	// 自动扩容逻辑
	//go autoScaling()

	// 优雅关闭
	handleShutdown()

	// 保持主协程运行
	select {}
}

func startConsumers(num int) {
	for i := 1; i <= num; i++ {
		c := fanout.NewConsumer(fmt.Sprintf("Consumer-%d", i))
		go c.Start()
	}
}

func autoScaling() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Adding new fanout...")
		startConsumers(1)
	}
}

func handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nInitiating graceful shutdown...")
		// 可以添加自定义清理逻辑
		os.Exit(0)
	}()
}
