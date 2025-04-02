// dlq_consumer.go
package main

import (
	"bessGin/config"
	"bessGin/util/rabbitmq/acksixin"
	"context"
	"log"
	"os/signal"
	"syscall"
)

func main() {

	cfg := config.Load()
	consumer := acksixin.NewDLXConsumer(cfg)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Fatalf("Consumer error: %v", err)
		}
	}()

	<-ctx.Done()
	consumer.Shutdown()
	log.Println("Shutdown completed")
}
