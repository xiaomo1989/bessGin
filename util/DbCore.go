package util

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"sync"
)

var (
	dsn string = "root:123456wan@tcp(localhost:3306)/hyperf?charset=utf8mb4"
)

var (
	dbInstance *gorm.DB
	onceCore   sync.Once
)

func DB() *gorm.DB {
	onceCore.Do(func() {
		var err error
		dbInstance, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   "",
				SingularTable: true,
			},
		})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
	})
	return dbInstance
}
