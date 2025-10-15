package article

import "time"

// Favorite 收藏表
// 注意：该功能暂未实现，仅保留数据库字段
type Favorite struct {
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	ArticleID uint      `gorm:"primaryKey;index" json:"article_id"`
	CreatedAt time.Time `json:"created_at"`
}
