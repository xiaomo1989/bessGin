package main

import (
	"bessGin/util"
	"fmt"
)

func main() {
	// 获取 RabbitMQ 单例
	rmq := util.GetRabbitMQInstance()
	msgs, err := rmq.Subscribe()
	if err != nil {
		fmt.Println("订阅失败", err)
		return
	}
	// 启动 Goroutine 处理 RabbitMQ 消息，避免阻塞 HTTP 请求

	for msg := range msgs {
		fmt.Println("收到 RabbitMQ 消息:", string(msg.Body))
		// 这里可以将消息存入数据库或日志
	}

	fmt.Println("成功")
}
