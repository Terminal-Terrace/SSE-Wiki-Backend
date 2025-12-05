package model

import (
	"terminal-terrace/sse-wiki/internal/model/article"
	"terminal-terrace/sse-wiki/internal/model/discussion"
	filemodel "terminal-terrace/sse-wiki/internal/model/file"
	"terminal-terrace/sse-wiki/internal/model/module"
	"terminal-terrace/sse-wiki/internal/model/user"

	"gorm.io/gorm"
)

// InitTable 自动迁移所有核心业务模型
func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(
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

		// 文件相关模型
		&filemodel.File{},
		&filemodel.ArticleVersionFile{},
	)
}
