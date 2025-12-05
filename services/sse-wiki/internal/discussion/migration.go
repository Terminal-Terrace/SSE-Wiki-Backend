package discussion

import (
	"gorm.io/gorm"

	discussionModel "terminal-terrace/sse-wiki/internal/model/discussion"
)

// MigrateDB 执行数据库迁移
// 在应用启动时调用此函数来自动创建/更新表结构
func MigrateDB(db *gorm.DB) error {
	return db.AutoMigrate(
		&discussionModel.Discussion{},
		&discussionModel.DiscussionComment{},
	)
}

// 如果需要在 main.go 或 database 包中统一管理迁移，可以这样使用：
/*
// 在 internal/database/migrate.go 中添加：

import "terminal-terrace/sse-wiki/internal/discussion"

func AutoMigrate(db *gorm.DB) error {
	// 迁移其他表...
	
	// 迁移讨论区表
	if err := discussion.MigrateDB(db); err != nil {
		return err
	}
	
	return nil
}
*/