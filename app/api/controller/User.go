package controller

import (
	userService "bessGin/app/service"
	"bessGin/util"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	ginAutoRouter "github.com/xiaomo1989/gin-auto-router"
	"io"
	"net/http"
	"strconv"
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

func (api *User) PushRabbit(c *gin.Context) {
	// 获取 RabbitMQ 单例
	rmq := util.GetRabbitMQInstance()
	defer rmq.Close()
	var req struct {
		Message string `json:"message"`
	}
	// 发布消息到 RabbitMQ
	err := rmq.Publish(req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "消息发布失败", "errrs": err})
	}
	c.JSON(http.StatusOK, gin.H{"message": "消息已发布"})
}

/*func (api *User) SubRabbit(c *gin.Context) {
	// 获取 RabbitMQ 单例
	rmq := util.GetRabbitMQInstance()
	defer rmq.Close()
	msgs, err := rmq.Subscribe()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "订阅失败", "errrs": err})
		return
	}
	// 监听 RabbitMQ 消息
	for msg := range msgs {
		fmt.Println("收到 RabbitMQ 消息:", string(msg.Body))
		c.JSON(http.StatusOK, gin.H{"message": string(msg.Body)})
	}
}
*/
/*
*
如果你的需求是 让后端后台处理 RabbitMQ 消息，并存入数据库/日志，可以使用 goroutine 处理消息，然后 HTTP 只返回 “订阅成功” 的响应。
优点：
不会阻塞 HTTP 请求，请求返回后，RabbitMQ 仍然在后台运行。
避免 c.JSON() 多次调用的问题。
适用于需要后台处理 RabbitMQ 消息的场景，例如写入数据库、日志等。
*/
func (api *User) SubRabbitGet(c *gin.Context) {
	// 获取 RabbitMQ 单例
	rmq := util.GetRabbitMQInstance()
	msgs, err := rmq.Subscribe()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "订阅失败", "details": err.Error()})
		return
	}
	// 启动 Goroutine 处理 RabbitMQ 消息，避免阻塞 HTTP 请求
	go func() {
		for msg := range msgs {
			fmt.Println("收到 RabbitMQ 消息:", string(msg.Body))
			// 这里可以将消息存入数据库或日志
		}
	}()

	// 立即返回 HTTP 响应，避免长时间阻塞
	c.JSON(http.StatusOK, gin.H{"message": "RabbitMQ 订阅已启动"})
}
