package dto

import (
	"encoding/json"
)

// StringSlice 自定义字符串切片类型，支持空字符串解析
type StringSlice []string

// UnmarshalJSON 实现自定义JSON解析，处理空字符串情况
func (s *StringSlice) UnmarshalJSON(data []byte) error {
	// 处理空字符串的情况
	if string(data) == `""` || string(data) == `null` {
		*s = []string{}
		return nil
	}

	// 正常解析数组
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*s = arr
	return nil
}

// CreateArticleRequest 创建文章请求
type CreateArticleRequest struct {
	Title            string      `json:"title" binding:"required,max=255"`
	ModuleID         uint        `json:"module_id" binding:"required"`
	Content          string      `json:"content" binding:"required"`
	CommitMessage    string      `json:"commit_message" binding:"required,max=255"`
	IsReviewRequired *bool       `json:"is_review_required"`
	Tags             StringSlice `json:"tags"`
}

// UpdateArticleBasicInfoRequest 更新文章基础信息请求
// 包含标题、标签、审核设置等不需要版本管理的基础属性
// 权限要求：admin、owner 或 moderator
type UpdateArticleBasicInfoRequest struct {
	Title            *string      `json:"title" binding:"omitempty,max=255"`
	Tags             *StringSlice `json:"tags"`
	IsReviewRequired *bool        `json:"is_review_required"`
}

// AddCollaboratorRequest 添加协作者请求
type AddCollaboratorRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=owner moderator"`
}

// SubmissionRequest 提交修改请求
// 注意：标签只能在创建文章时设置，更新时不支持修改标签
type SubmissionRequest struct {
	Content       string `json:"content" binding:"required"`
	CommitMessage string `json:"commit_message" binding:"required,max=255"`
	BaseVersionID uint   `json:"base_version_id" binding:"required"`
}

// ReviewActionRequest 审核操作请求
type ReviewActionRequest struct {
	Action        string  `json:"action" binding:"required,oneof=approve reject"`
	Notes         string  `json:"notes"`
	MergedContent *string `json:"merged_content"` // 仅当手动解决冲突时需要
}

// ArticleResponse 文章响应
type ArticleResponse struct {
	ID               uint     `json:"id"`
	Title            string   `json:"title"`
	ModuleID         uint     `json:"module_id"`
	CurrentVersionID *uint    `json:"current_version_id"`
	CurrentUserRole  string   `json:"current_user_role"`
	IsReviewRequired *bool    `json:"is_review_required"`
	ViewCount        uint     `json:"view_count"`
	Tags             []string `json:"tags"`
	CreatedBy        uint     `json:"created_by"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

// VersionResponse 版本响应
type VersionResponse struct {
	ID            uint   `json:"id"`
	ArticleID     uint   `json:"article_id"`
	VersionNumber int    `json:"version_number"`
	Content       string `json:"content"`
	CommitMessage string `json:"commit_message"`
	AuthorID      uint   `json:"author_id"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

// SubmissionResponse 提交响应
type SubmissionResponse struct {
	ID                uint   `json:"id"`
	ArticleID         uint   `json:"article_id"`
	ArticleTitle      string `json:"article_title"`
	ProposedVersionID uint   `json:"proposed_version_id"`
	BaseVersionID     uint   `json:"base_version_id"`
	SubmittedBy       uint   `json:"submitted_by"`
	SubmittedByName   string `json:"submitted_by_name"`
	ReviewedBy        *uint  `json:"reviewed_by,omitempty"`
	Status            string `json:"status"`
	CommitMessage     string `json:"commit_message"`
	HasConflict       bool   `json:"has_conflict"`
	// TODO: AI评分功能 - AI评分字段，后续接入AI审核服务时使用
	AIScore    *int   `json:"ai_score,omitempty"`
	CreatedAt  string `json:"created_at"`
	ReviewedAt string `json:"reviewed_at,omitempty"`
}

// ConflictData 冲突数据
type ConflictData struct {
	BaseContent          string `json:"base_content"`
	TheirContent         string `json:"their_content"`
	OurContent           string `json:"our_content"`
	HasConflict          bool   `json:"has_conflict"`
	BaseVersionNumber    int    `json:"base_version_number"`
	CurrentVersionNumber int    `json:"current_version_number"`
	SubmitterName        string `json:"submitter_name"`
}

// ArticleListItem 文章列表项
type ArticleListItem struct {
	ID               uint     `json:"id"`
	Title            string   `json:"title"`
	ModuleID         uint     `json:"module_id"`
	CurrentVersionID *uint    `json:"current_version_id"`
	ViewCount        uint     `json:"view_count"`
	Tags             []string `json:"tags"`
	CreatedBy        uint     `json:"created_by"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

// ArticleListResponse 文章列表响应（分页）
type ArticleListResponse struct {
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Articles []ArticleListItem `json:"articles"`
}
