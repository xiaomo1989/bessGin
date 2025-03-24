package controller

import (
	"bessGin/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	ginAutoRouter "github.com/xiaomo1989/gin-auto-router"
	"net/http"
	"time"
)

func init() {
	ginAutoRouter.Register(&Article{})
}

type Article struct{}

func (api *Article) ListGet(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "ok",
		"data": "Article:List",
	})
}

func (api *Article) Test(c *gin.Context) {
	/*fmt.Println("6666666666666666666")
	rdb, ctx := util.GetRedis()
	err := rdb.Set(ctx, "name", "Gin+Redis", 10*time.Second).Err()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		//return
	}*/

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "ok",
		"data": "我是谁",
	})
}

func (api *Article) TestGet(c *gin.Context) {
	err := util.GetRedis().Set("name", "Gin+Redis", 10*time.Second)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}

	value, err1 := util.GetRedis().Get("name")
	if err1 == redis.Nil {
		c.JSON(404, gin.H{"message": "数据不存在"})
	} else if err1 != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}
	fmt.Println(value)

	c.JSON(http.StatusOK, gin.H{
		"code":  1,
		"msg":   "数据已存入 Redis",
		"value": value,
		"data":  "Article:Test",
	})
}
