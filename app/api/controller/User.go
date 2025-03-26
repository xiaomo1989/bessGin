package controller

import (
	userService "bessGin/app/service"
	"bessGin/util"
	"bessGin/util/rabbitmq/fanout"
	"encoding/json"
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
	message := c.PostForm("message")
	// 初始化RabbitMQ连接
	fanout.GetConnection()
	//发消息
	err := fanout.Publish([]byte(message))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish message"})
	}
	c.JSON(http.StatusOK, gin.H{"message": "消息已发布"})
}
