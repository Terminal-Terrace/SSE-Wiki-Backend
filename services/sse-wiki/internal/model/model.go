package model

import (
	"gorm.io/gorm"
	"terminal-terrace/sse-wiki/internal/model/article"
	"terminal-terrace/sse-wiki/internal/model/discussion"
	"terminal-terrace/sse-wiki/internal/model/module"
	"terminal-terrace/sse-wiki/internal/model/user"
)

func InitTable(db *gorm.DB) error {
	// 自动迁移数据库表结构
	err := db.AutoMigrate(
		// 用户模型
		&user.User{},

		// 模块相关模型
		&module.Module{},
		&module.ModuleModerator{},

		// 文章相关模型
		&article.Article{},
		&article.ArticleCollaborator{},
		&article.ArticleVersion{},
		&article.ReviewSubmission{},
		&article.VersionConflict{},
		&article.ArticleReference{},
		&article.Tag{},
		&article.ArticleTag{},
		&article.Favorite{},

		// 讨论相关模型
		&discussion.Discussion{},
		&discussion.DiscussionComment{},
	)
	if err != nil {
		return err
	}
	return nil
}