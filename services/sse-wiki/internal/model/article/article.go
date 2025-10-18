// Package article 文章相关模型
package article

import (
	"time"
)

// Article 文章基础信息表
type Article struct {
	ID uint `gorm:"primaryKey" json:"id"`
	// 指向当前被发布、对外可见的版本ID
	CurrentVersionID *uint  `gorm:"index" json:"current_version_id"`
	Title            string `gorm:"type:varchar(255);not null" json:"title"`
	// 自定义模块的外键
	ModuleID  uint      `gorm:"not null;index" json:"module_id"`
	CreatedBy uint      `gorm:"not null;index" json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// 当前文章是否需要审核（默认true，可由管理员设置）
	IsReviewRequired bool `gorm:"default:true" json:"is_review_required"`
	// 阅读量统计
	ViewCount uint `gorm:"default:0" json:"view_count"`
}

// ArticleCollaborator 文章协作者表（权限独立设计）
// 注意：文章权限完全独立，不继承模块权限，需要单独设置
// 权限层级：全局admin > owner > moderator > 普通用户
// 普通用户（不在协作者表）默认可以提交修改，无需额外角色
type ArticleCollaborator struct {
	ArticleID uint `gorm:"primaryKey" json:"article_id"`
	UserID    uint `gorm:"primaryKey" json:"user_id"`
	// 角色: owner(作者，可删除、设置审核、审核、添加协作者)
	//      moderator(审核员，可审核)
	// 注意：
	// 1. 普通用户不需要在此表中，默认有提交权限
	// 2. 全局admin（来自认证系统）拥有所有权限
	Role      string    `gorm:"type:varchar(50);not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ArticleVersion 文章版本历史表 (全量存储)
type ArticleVersion struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	ArticleID uint `gorm:"not null;uniqueIndex:idx_article_version_unique" json:"article_id"`
	// 版本号，在article_id下递增 (1, 2, 3...)
	VersionNumber int `gorm:"not null;uniqueIndex:idx_article_version_unique" json:"version_number"`
	// Markdown原文，全量存储
	Content string `gorm:"type:text;not null" json:"content"`
	// 提交信息
	CommitMessage string `gorm:"type:varchar(255)" json:"commit_message"`
	// 版本作者ID
	AuthorID uint `gorm:"not null;index" json:"author_id"`
	// 版本状态: pending(待审核/未发布), published(已发布), rejected(已驳回)
	Status string `gorm:"type:varchar(50);default:'pending'" json:"status"`
	// 基于哪个版本创建（用于diff对比，v1为null）
	BaseVersionID *uint `gorm:"index" json:"base_version_id,omitempty"`
	// 审核合并时所针对的当前线上版本ID（无冲突为Base，否则为当前版本）
	MergedAgainstVersionID *uint     `gorm:"index" json:"merged_against_version_id,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
}

// ReviewSubmission 审核提交表 (等同于Pull Request)
type ReviewSubmission struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	ArticleID uint `gorm:"not null;index:idx_article_status" json:"article_id"`
	// 本次提交创建的新版本ID（提交者的修改内容）
	ProposedVersionID uint `gorm:"not null" json:"proposed_version_id"`
	// 本次提交所基于的版本ID（用于3路合并的base）
	BaseVersionID uint `gorm:"not null" json:"base_version_id"`
	// 提交时的标签（JSON数组，审核通过后才应用到article_tags表）
	ProposedTags string `gorm:"type:json" json:"proposed_tags,omitempty"`
	// TODO: AI评分功能 - AI评分 (1-100)，后续接入AI审核服务时使用
	AIScore *int `gorm:"type:smallint" json:"ai_score,omitempty"`
	// TODO: AI评分功能 - AI审核建议，后续接入AI审核服务时使用
	AISuggestions string `gorm:"type:text" json:"ai_suggestions,omitempty"`
	// 提交人ID
	SubmittedBy uint `gorm:"not null;index" json:"submitted_by"`
	// 审核人ID
	ReviewedBy *uint `gorm:"index" json:"reviewed_by,omitempty"`
	// 状态: pending, approved, rejected, conflict_detected, merged
	Status string `gorm:"type:varchar(50);default:'pending';index:idx_article_status" json:"status"`
	// 审核备注（审核人填写）
	ReviewNotes string `gorm:"type:text" json:"review_notes,omitempty"`
	// 3路合并结果（如果有冲突，存储带冲突标记的内容）
	MergeResult string `gorm:"type:text" json:"merge_result,omitempty"`
	// 是否有冲突
	HasConflict bool `gorm:"default:false" json:"has_conflict"`
	// 审核或冲突检测时参考的线上版本ID
	MergedAgainstVersionID *uint      `gorm:"index" json:"merged_against_version_id,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	ReviewedAt             *time.Time `json:"reviewed_at,omitempty"`
}

// VersionConflict 版本冲突记录表 (用于追踪和解决冲突)
type VersionConflict struct {
	ID uint `gorm:"primaryKey" json:"id"`
	// 发生冲突的提交ID
	SubmissionID uint `gorm:"not null;index" json:"submission_id"`
	// 与哪个已发布的版本发生了冲突
	ConflictWithVersionID uint `gorm:"not null" json:"conflict_with_version_id"`
	// 状态: detected, resolved
	Status string `gorm:"type:varchar(50);default:'detected'" json:"status"`
	// 解决后的最终版本ID
	ResolvedVersionID *uint `json:"resolved_version_id,omitempty"`
	// 解决冲突的用户ID（通常是管理员）
	ResolvedBy *uint `json:"resolved_by,omitempty"`
	// 冲突的具体位置和内容（JSON格式存储）
	ConflictDetails string     `gorm:"type:text" json:"conflict_details,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
}
