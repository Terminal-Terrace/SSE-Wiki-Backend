// Package discussion 讨论功能相关模型
package discussion

import (
	"time"

	"gorm.io/gorm"
)

// Discussion 讨论主题表
// 每篇文章对应一个讨论区
type Discussion struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ArticleID   uint           `gorm:"not null;uniqueIndex;comment:文章ID" json:"article_id"`
	Title       string         `gorm:"type:varchar(255);not null;comment:讨论标题" json:"title"`
	Description string         `gorm:"type:text;comment:讨论描述" json:"description"`
	CreatedBy   uint           `gorm:"not null;index;comment:创建者ID" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"` // 软删除

	// 关联
	Comments []DiscussionComment `gorm:"foreignKey:DiscussionID" json:"comments,omitempty"`
}

// TableName 指定表名
func (Discussion) TableName() string {
	return "discussions"
}

// DiscussionComment 讨论评论表
// 支持多级嵌套回复
type DiscussionComment struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	DiscussionID uint           `gorm:"not null;index;comment:讨论ID" json:"discussion_id"`
	ParentID     *uint          `gorm:"index;comment:父评论ID，NULL表示顶级评论" json:"parent_id,omitempty"`
	Content      string         `gorm:"type:text;not null;comment:评论内容" json:"content"`
	CreatedBy    uint           `gorm:"not null;index;comment:创建者ID" json:"created_by"`
	CreatedAt    time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updated_at"`

	IsDeleted bool 				`gorm:"column:is_deleted;default:false"`

	// 关联（可选，用于预加载）
	Discussion *Discussion           `gorm:"foreignKey:DiscussionID" json:"-"`
	Parent     *DiscussionComment    `gorm:"foreignKey:ParentID" json:"-"`
	Replies    []DiscussionComment   `gorm:"foreignKey:ParentID" json:"-"`
	Creator    interface{}           `gorm:"-" json:"creator,omitempty"` // 不映射到数据库，用于返回用户信息
}

// TableName 指定表名
func (DiscussionComment) TableName() string {
	return "discussion_comments"
}

// BeforeCreate GORM钩子：创建前的验证
func (c *DiscussionComment) BeforeCreate(tx *gorm.DB) error {
	// 可以在这里添加额外的验证逻辑
	if c.Content == "" {
		return gorm.ErrInvalidData
	}
	return nil
}
