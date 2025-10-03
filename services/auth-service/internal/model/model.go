package model

import (
	"gorm.io/gorm"
	"terminal-terrace/auth-service/internal/model/user"
	"terminal-terrace/auth-service/internal/model/user_providers"
)

func InitTable(db *gorm.DB) error {
	// 自动迁移数据库表结构
	err := db.AutoMigrate(
		// TODO: 在这里添加模型
		&user.User{},
		&userproviders.UserProvider{},
	)
	if err != nil {
		return err
	}
	return nil
}
