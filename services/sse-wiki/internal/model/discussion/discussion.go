// Package discussion 讨论功能相关模型
package discussion

import "time"

// Discussion 讨论主题表
type Discussion struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ArticleID   uint      `gorm:"not null;index" json:"article_id"`
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedBy   uint      `gorm:"not null;index" json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DiscussionComment 讨论评论表（支持多级嵌套）
type DiscussionComment struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DiscussionID uint      `gorm:"not null;index" json:"discussion_id"`
	ParentID     *uint     `gorm:"index" json:"parent_id,omitempty"` // NULL表示顶级评论
	Content      string    `gorm:"type:text;not null" json:"content"`
	CreatedBy    uint      `gorm:"not null;index" json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
