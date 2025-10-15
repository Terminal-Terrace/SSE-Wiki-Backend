package article

import "time"

// ArticleReference 文章引用关系表（多对多）
// 支持复杂的引用网络，不再使用prev/next的线性关系
type ArticleReference struct {
	FromArticleID uint      `gorm:"primaryKey;index" json:"from_article_id"` // 引用源
	ToArticleID   uint      `gorm:"primaryKey;index" json:"to_article_id"`   // 被引用
	// 引用类型：prerequisite(前置知识), related(相关文章), extends(扩展阅读)
	ReferenceType string    `gorm:"type:varchar(50);not null" json:"reference_type"`
	CreatedAt     time.Time `json:"created_at"`
}
