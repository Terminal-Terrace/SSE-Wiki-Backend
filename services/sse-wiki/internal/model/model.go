package model

import (
	"gorm.io/gorm"
)

func InitTable(db *gorm.DB) error {
	// 自动迁移数据库表结构
	err := db.AutoMigrate(
		// TODO: 在这里添加模型
	)
	if err != nil {
		return err
	}
	return nil
}