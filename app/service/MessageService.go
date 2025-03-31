package service

import (
	model "bessGin/app/models"
	"bessGin/util"
	"time"
)

type MessageService struct{}

func (r *MessageService) SaveRecord() error {
	format := time.Now().Format("2006-01-02 15:04:05")
	newUser := model.MessageRecord{Content: "我是消息内容", ConsumerID: "123", ProcessedAt: format}
	_ = util.DB().Create(&newUser)
	return nil
}

func (r *MessageService) Exists(msgID string) (bool, error) {
	return true, nil
}
