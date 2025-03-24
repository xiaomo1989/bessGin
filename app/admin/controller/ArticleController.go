package controller

import (
	"github.com/gin-gonic/gin"
	ginAutoRouter "github.com/xiaomo1989/gin-auto-router"
	"net/http"
)

func init() {
	ginAutoRouter.Register(&ArticleController{})
}

type ArticleController struct{}

func (api *ArticleController) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "ok",
		"data": "Article:List",
	})
}

func (api *ArticleController) Test(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "ok",
		"data": "Article:Test",
	})
}
func (api *ArticleController) TestGet(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "ok",
		"data": "Article:Test",
	})
}
