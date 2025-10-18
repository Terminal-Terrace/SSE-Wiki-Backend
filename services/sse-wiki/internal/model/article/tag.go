package article

import "time"

// Tag 标签表
type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	Color     string    `gorm:"type:varchar(20);default:'#3b82f6'" json:"color"` // 标签颜色
	CreatedAt time.Time `json:"created_at"`
}

// ArticleTag 文章-标签关联表
type ArticleTag struct {
	ArticleID uint      `gorm:"primaryKey;index" json:"article_id"`
	TagID     uint      `gorm:"primaryKey;index" json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
}
