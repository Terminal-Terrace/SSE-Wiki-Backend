package discussion

import (
	"time"
	discussionModel "terminal-terrace/sse-wiki/internal/model/discussion"
)

// ========== 请求 DTO ==========

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=5000"` // 评论内容，1-5000字符
}

// UpdateCommentRequest 更新评论请求
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

// ========== 响应 DTO ==========

// CommentResponse 评论响应（树状结构）
type CommentResponse struct {
	ID           uint              `json:"id"`
	DiscussionID uint              `json:"discussion_id"`
	ParentID     *uint             `json:"parent_id,omitempty"`
	Content      string            `json:"content"`
	CreatedBy    uint              `json:"created_by"`
	Creator      *UserInfo         `json:"creator,omitempty"` // 评论创建者信息
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	
	Replies      []*CommentResponse `json:"replies,omitempty"` // 子评论（递归结构）
	
	ReplyCount   int                `json:"reply_count"`       // 直接回复数量

	IsDeleted  bool                 `json:"is_deleted"`
}

// UserInfo 用户信息（简化版）
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar,omitempty"`
}

// DiscussionResponse 讨论区响应
type DiscussionResponse struct {
	ID          uint      `json:"id"`
	ArticleID   uint      `json:"article_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CommentsListResponse 评论列表响应（包含讨论区信息）
type CommentsListResponse struct {
	Discussion *DiscussionResponse `json:"discussion,omitempty"`
	Comments   []*CommentResponse   `json:"comments"`
	Total      int                 `json:"total"` // 总评论数（包括所有层级）
}

// ========== 通用响应 ==========

// Response 标准API响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ========== 辅助函数：Model -> DTO 转换 ==========

// ToCommentResponse 将 Model 转换为 Response DTO（不包含子评论）
func ToCommentResponse(comment *discussionModel.DiscussionComment) *CommentResponse {
	createdAt := comment.CreatedAt
	if createdAt.IsZero() {
		// 如果时间为零值（0001-01-01），则使用当前时间作为备用，
		// 这样前端会收到一个有效时间字符串，而不是 null/undefined。
		// 这通常在评论创建后立即返回DTO时可能会发生。
		createdAt = time.Now() 
	}

	resp := &CommentResponse{
		ID:           comment.ID,
		DiscussionID: comment.DiscussionID,
		ParentID:     comment.ParentID,
		Content:      comment.Content,
		CreatedBy:    comment.CreatedBy,
		CreatedAt:    createdAt,
		UpdatedAt:    comment.UpdatedAt,
		Replies:      []*CommentResponse{},
		ReplyCount:   0,
		IsDeleted:    comment.IsDeleted,
	}

	// 如果有关联的用户信息
	if comment.Creator != nil {
		if userInfo, ok := comment.Creator.(*UserInfo); ok {
			resp.Creator = userInfo
		}
	}

	if comment.IsDeleted {
        resp.Content = "该评论已被删除"
        resp.Creator = nil // 可选：不显示作者信息
    }

	return resp
}

// ToDiscussionResponse 将 Discussion Model 转换为 Response DTO
func ToDiscussionResponse(discussion *discussionModel.Discussion) *DiscussionResponse {
	if discussion == nil {
		return nil
	}

	return &DiscussionResponse{
		ID:          discussion.ID,
		ArticleID:   discussion.ArticleID,
		Title:       discussion.Title,
		Description: discussion.Description,
		CreatedBy:   discussion.CreatedBy,
		CreatedAt:   discussion.CreatedAt,
		UpdatedAt:   discussion.UpdatedAt,
	}
}