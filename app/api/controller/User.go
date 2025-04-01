package controller

import (
	"bessGin/app/models"
	userService "bessGin/app/service"
	"bessGin/config"
	"bessGin/util"
	ackRabbit "bessGin/util/rabbitmq/acksixin"
	"bessGin/util/rabbitmq/fanout"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	ginAutoRouter "github.com/xiaomo1989/gin-auto-router"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func init() {
	ginAutoRouter.Register(&User{})
}

type User struct {
	Id   int
	Name string
	Sex  int
}

func decodeJSON(r io.Reader, obj any) error {
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}

// 分页
func (api *User) Pages(c *gin.Context) {
	/*var data User
	// 将c.Request.Body中的数据解析为data
	err := decodeJSON(c.Request.Body, &data)*/
	condMap := make(map[string]interface{})
	userName := c.Query("name")
	id := c.Query("id")
	//userName := c.PostForm("name")
	if userName != "" {
		condMap["name"] = userName
	}
	if id != "" {
		condMap["id"] = id
	}
	page := c.DefaultQuery("page", "1")
	pageIndex, err := strconv.Atoi(page)
	if err != nil {
		panic(err)
	}
	pageSize := 2
	users := userService.UserPage(condMap, pageIndex, pageSize)
	total := userService.UserTotal(condMap)
	pages := util.Paginator(pageIndex, pageSize, total)
	c.JSON(http.StatusOK, gin.H{
		"code":  1,
		"msg":   "ok",
		"data":  users,
		"pages": pages,
		"total": total,
	})
}

// 列表 带传参 可根据条件查询
func (api *User) ListGet(c *gin.Context) {
	userName := c.Query("name")
	condMap := make(map[string]interface{})
	if userName != "" {
		condMap["name"] = userName
	}
	list := userService.UserList(condMap)
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "ok",
		"data": list,
	})
}

// 详细信息
func (api *User) UserInfo(c *gin.Context) {
	userName := c.Query("name")
	//condMap := make(map[string]interface{})
	id := c.Query("id")
	condMap := map[string]interface{}{
		"name": userName, // 进行 LIKE 查询
		"id":   id,       // 进行精确匹配
	}

	list := userService.UserInfo(condMap)
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "ok",
		"data": list,
	})
}

func (api *User) TestGet(c *gin.Context) {
	test(&testing.B{})
}

func test(b *testing.B) {
	var demoMap map[string]string = map[string]string{
		"a": "a",
		"b": "b",
	}
	// 模拟并发写map
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			demoMap["a"] = "aa"
		}
	})
}

// 发布消息
func (api *User) UserInfo1(c *gin.Context) {
	message := "Hello, Redis Pub/Sub!" // 示例消息
	channel := "mychannel"             // 频道
	// 获取 Redis 单例客户端
	redisClient := util.GetRedis()
	// 发布消息
	err := redisClient.Publish(channel, message)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "ok",
			"data": err,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": "success",
	})
}

// 生产消息
func (api *User) PushRabbit(c *gin.Context) {
	message := c.PostForm("message")
	// 初始化RabbitMQ连接
	fanout.GetConnection()
	//发消息
	err := fanout.PublishMessage(message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish message"})
	}
	c.JSON(http.StatusOK, gin.H{"message": "消息已发布"})
}

// 手动确认机制+死信队列
func (api *User) PushSi(c *gin.Context) {
	// 加载配置
	cfg := config.Load()

	// 初始化生产者
	producer, err := ackRabbit.NewProducer(cfg)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("Producer shutdown error: %v", err)
		}
	}()

	// 优雅关闭处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("\nReceived shutdown signal")
		cancel()
	}()
	// 启动生产循环
	produceOrders(ctx, producer)
}

func produceOrders(ctx context.Context, p *ackRabbit.Producer) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for i := 1; ; i++ {
		select {
		case <-ctx.Done():
			log.Println("Stopping order production")
			return
		case <-ticker.C:
			order := generateOrder(i)
			if err := p.PublishOrder(ctx, order); err != nil {
				log.Printf("Failed to publish order %d: %v", i, err)
				continue
			}
			log.Printf("Successfully published order %s", order.ID)
		}
	}
}

func generateOrder(seq int) models.Order {
	return models.Order{
		ID:        fmt.Sprintf("ORD-%d-%d", time.Now().Unix(), seq),
		Product:   fmt.Sprintf("Product-%d", rand.Intn(10)+1),
		Quantity:  rand.Intn(10) + 1,
		Timestamp: time.Now().UTC(),
	}
}
