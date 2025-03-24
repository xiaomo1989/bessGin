package controller

import (
	userService "bessGin/app/service"
	"bessGin/util"
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
