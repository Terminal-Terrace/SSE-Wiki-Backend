package discussion

import (
	"gorm.io/gorm"

	discussionModel "terminal-terrace/sse-wiki/internal/model/discussion"
)

// DiscussionRepository 讨论区数据访问接口
type DiscussionRepository interface {
	// Discussion 相关
    FindDiscussionByArticleID(articleID uint) (*discussionModel.Discussion, error)
    CreateDiscussion(discussion *discussionModel.Discussion) error

	// Comment 相关
    FindCommentByID(commentID uint) (*discussionModel.DiscussionComment, error)
    FindCommentsByDiscussionID(discussionID uint) ([]discussionModel.DiscussionComment, error)
    CreateComment(comment *discussionModel.DiscussionComment) error
    UpdateComment(comment *discussionModel.DiscussionComment) error
    DeleteComment(commentID uint) error
    CountCommentsByDiscussionID(discussionID uint) (int64, error)
}

// discussionRepository 实现
type discussionRepository struct {
	db *gorm.DB
}

// NewDiscussionRepository 创建 Repository 实例
func NewDiscussionRepository(db *gorm.DB) DiscussionRepository {
	return &discussionRepository{db: db}
}

// ========== Discussion 相关操作 ==========

// FindDiscussionByArticleID 根据文章ID查找讨论区
func (r *discussionRepository) FindDiscussionByArticleID(articleID uint) (*discussionModel.Discussion, error) {
    var discussion discussionModel.Discussion
	err := r.db.Where("article_id = ?", articleID).First(&discussion).Error
	if err != nil {
		return nil, err
	}
	return &discussion, nil
}

// CreateDiscussion 创建讨论区
func (r *discussionRepository) CreateDiscussion(discussion *discussionModel.Discussion) error {
	return r.db.Create(discussion).Error
}

// ========== Comment 相关操作 ==========

// FindCommentByID 根据ID查找评论
func (r *discussionRepository) FindCommentByID(commentID uint) (*discussionModel.DiscussionComment, error) {
	var comment discussionModel.DiscussionComment
	err := r.db.First(&comment, commentID).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// FindCommentsByDiscussionID 获取讨论区的所有评论（扁平列表）
func (r *discussionRepository) FindCommentsByDiscussionID(discussionID uint) ([]discussionModel.DiscussionComment, error) {
    var comments []discussionModel.DiscussionComment
	err := r.db.Where("discussion_id = ?", discussionID).
		Order("created_at ASC"). // 按创建时间升序
		Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// CreateComment 创建评论
func (r *discussionRepository) CreateComment(comment *discussionModel.DiscussionComment) error {
	return r.db.Create(comment).Error
}

// UpdateComment 更新评论
func (r *discussionRepository) UpdateComment(comment *discussionModel.DiscussionComment) error {
	// 只更新 content 和 updated_at
	return r.db.Model(comment).Updates(map[string]interface{}{
		"content":    comment.Content,
		"updated_at": comment.UpdatedAt,
	}).Error
}

// DeleteComment 标记评论为已删除（逻辑删除，设置 is_deleted=true）
func (r *discussionRepository) DeleteComment(commentID uint) error {
    return r.db.Model(&discussionModel.DiscussionComment{}).
        Where("id = ?", commentID).
        Updates(map[string]interface{}{"is_deleted": true}).Error
}

// CountCommentsByDiscussionID 统计讨论区的评论总数
func (r *discussionRepository) CountCommentsByDiscussionID(discussionID uint) (int64, error) {
	var count int64
	err := r.db.Model(&discussionModel.DiscussionComment{}).
		Where("discussion_id = ?", discussionID).
		Count(&count).Error
	return count, err
}