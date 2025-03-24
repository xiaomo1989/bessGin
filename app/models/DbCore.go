package models

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
	once       sync.Once
)

func DB() *gorm.DB {
	once.Do(func() {
		var err error
		dbInstance, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   "tp_",
				SingularTable: true,
			},
		})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
	})
	return dbInstance
}
