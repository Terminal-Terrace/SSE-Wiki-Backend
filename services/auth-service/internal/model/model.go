package model

import (
	"fmt"
	"terminal-terrace/auth-service/internal/model/user"
	userproviders "terminal-terrace/auth-service/internal/model/user_providers"

	"gorm.io/gorm"
)

// GetModels 返回所有需要迁移的模型
func GetModels() []interface{} {
	return []interface{}{
		&user.User{},
		&userproviders.UserProvider{},
	}
}

func InitTable(db *gorm.DB) error {
	models := GetModels()

	// 执行自动迁移
	err := db.AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("数据库表迁移失败: %v", err)
	}

	return nil
}
