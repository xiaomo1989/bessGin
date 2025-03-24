package controller

import (
	"github.com/gin-gonic/gin"
	ginAutoRouter "github.com/xiaomo1989/gin-auto-router"
	"net/http"
)

func init() {
	ginAutoRouter.Register(&Article{})
}

type Article struct{}

func (api *Article) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "ok",
		"data": "Article:List",
	})
}

func (api *Article) Test(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "ok",
		"data": "Article:Test1",
	})
}

func (api *Article) TestGet(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "ok",
		"data": "Article:Test",
	})
}
