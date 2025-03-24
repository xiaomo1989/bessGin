package service

import (
	model "bessGin/app/models"
	"fmt"
	"strings"
)

// 列表方法
// gorm的Unscoped方法设置tx.Statement.Unscoped为true；针对软删除会追加SoftDeleteDeleteClause，即设置deleted_at为指定的时间戳；而callbacks的Delete方法在db.Statement.Unscoped为false的时候才追加db.Statement.Schema.DeleteClauses，而Unscoped则执行的是物理删除。
func UserList(wheres map[string]interface{}) (list []model.User) {
	var users []model.User
	model.DB().Debug().Where(wheres).Unscoped().Order("id desc").Find(&users)
	return users
}

// 分页方法
func UserPage(wheres map[string]interface{}, pageIndex int, pageSize int) (list []model.User) {
	offset := pageSize * (pageIndex - 1)
	model.DB().Debug().Where(wheres).Unscoped().Order("id desc").Limit(pageSize).Offset(offset).Find(&list)
	return list
}

// 取得总行数
func UserTotal(wheres map[string]interface{}) int64 {
	var user []model.User
	var total int64
	model.DB().Debug().Where(wheres).Find(&user).Count(&total)
	return total
}

func UserInfo(wheres map[string]interface{}) model.User {
	var user model.User
	db := model.DB().Debug()

	for key, value := range wheres {
		if key == "name" {
			if name, ok := value.(string); ok {
				name = strings.TrimSpace(name)
				db = db.Where("name LIKE ?", "%"+name+"%")
			}
		} else {
			db = db.Where(fmt.Sprintf("%s = ?", key), value)
		}
	}

	db.Order("id desc").First(&user)
	return user
}
