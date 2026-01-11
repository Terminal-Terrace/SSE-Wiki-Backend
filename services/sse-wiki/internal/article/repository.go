package article

import (
	"time"

	"terminal-terrace/sse-wiki/internal/model/article"

	"gorm.io/gorm"
)

// ArticleRepository 文章仓储层
type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

// ===== Article 基础操作 =====

func (r *ArticleRepository) GetByID(id uint) (*article.Article, error) {
	var art article.Article
	err := r.db.First(&art, id).Error
	return &art, err
}

func (r *ArticleRepository) GetFavoriteByUserId(userID uint) ([]uint32, error) {
	var articleIDs []uint32
	err := r.db.Model(&article.Favorite{}).
		Where("user_id = ?", userID).
		Pluck("article_id", &articleIDs).Error
	return articleIDs, err
}

func (r *ArticleRepository) UpdateUserFavourite(userId uint, articleId uint, is_added bool) (string, error) {
	fav := article.Favorite{UserID: userId, ArticleID: articleId}
	if is_added {
		// Add to favorites
		// Check if it already exists, if not, create it.
		result := r.db.FirstOrCreate(&fav, fav)
		if result.Error != nil {
			return "", result.Error
		}
		if result.RowsAffected > 0 {
			return "add successfully", nil
		}
		return "already exists", nil
	} else {
		// Remove from favorites
		result := r.db.Where("user_id = ? AND article_id = ?", userId, articleId).Delete(&article.Favorite{})
		if result.Error != nil {
			return "", result.Error
		}
		if result.RowsAffected > 0 {
			return "removed successfully", nil
		}
		return "not found", nil
	}
}

func (r *ArticleRepository) Create(art *article.Article) error {
	return r.db.Create(art).Error
}

func (r *ArticleRepository) Update(art *article.Article) error {
	return r.db.Save(art).Error
}

func (r *ArticleRepository) Delete(id uint) error {
	return r.db.Delete(&article.Article{}, id).Error
}

// CheckPermission 检查用户对文章的权限
// 权限层级：全局admin > owner > moderator > 普通用户
// userRole: 全局角色（来自认证系统），"admin" 表示系统管理员
// requiredRole: 需要的文章协作者角色（"owner" 或 "moderator"）
func (r *ArticleRepository) CheckPermission(articleID uint, userID uint, userRole string, requiredRole string) bool {
	// 1. 全局admin拥有所有权限
	if userRole == "admin" {
		return true
	}

	// 2. 检查文章协作者表
	var collaborator article.ArticleCollaborator
	err := r.db.Where("article_id = ? AND user_id = ?", articleID, userID).
		First(&collaborator).Error

	if err != nil {
		// 用户不在协作者表中，没有特殊权限
		return false
	}

	// 3. 角色权限层级：admin > moderator
	roleLevel := map[string]int{
		"moderator": 1,
		"admin":     2,
	}

	return roleLevel[collaborator.Role] >= roleLevel[requiredRole]
}

// GetUserRole 获取用户在文章中的角色
func (r *ArticleRepository) GetUserRole(articleID uint, userID uint) *string {
	var collaborator article.ArticleCollaborator
	err := r.db.Where("article_id = ? AND user_id = ?", articleID, userID).
		First(&collaborator).Error

	if err != nil {
		return nil
	}

	return &collaborator.Role
}

// IncrementViewCount 增加阅读量
func (r *ArticleRepository) IncrementViewCount(articleID uint) error {
	return r.db.Model(&article.Article{}).
		Where("id = ?", articleID).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

// AddCollaborator 添加协作者
func (r *ArticleRepository) AddCollaborator(articleID uint, userID uint, role string) error {
	collaborator := &article.ArticleCollaborator{
		ArticleID: articleID,
		UserID:    userID,
		Role:      role,
		CreatedAt: time.Now(),
	}
	return r.db.Save(collaborator).Error
}

// RemoveCollaborator 移除协作者
func (r *ArticleRepository) RemoveCollaborator(articleID uint, userID uint) error {
	return r.db.Where("article_id = ? AND user_id = ?", articleID, userID).
		Delete(&article.ArticleCollaborator{}).Error
}

// GetCollaborators 获取文章的所有协作者
func (r *ArticleRepository) GetCollaborators(articleID uint) ([]article.ArticleCollaborator, error) {
	var collaborators []article.ArticleCollaborator
	err := r.db.Where("article_id = ?", articleID).Find(&collaborators).Error
	return collaborators, err
}

// ListByModuleID 根据模块ID获取文章列表
func (r *ArticleRepository) ListByModuleID(moduleID uint, offset, limit int) ([]article.Article, int64, error) {
	var articles []article.Article
	var total int64

	query := r.db.Model(&article.Article{}).Where("module_id = ?", moduleID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := query.Offset(offset).Limit(limit).Order("updated_at DESC").Find(&articles).Error
	return articles, total, err
}

// VersionRepository 版本仓储层
type VersionRepository struct {
	db *gorm.DB
}

func NewVersionRepository(db *gorm.DB) *VersionRepository {
	return &VersionRepository{db: db}
}

func (r *VersionRepository) Create(version *article.ArticleVersion) error {
	return r.db.Create(version).Error
}

func (r *VersionRepository) GetByID(id uint) (*article.ArticleVersion, error) {
	var version article.ArticleVersion
	err := r.db.First(&version, id).Error
	return &version, err
}

func (r *VersionRepository) GetContent(versionID uint) (string, error) {
	var version article.ArticleVersion
	err := r.db.Select("content").First(&version, versionID).Error
	return version.Content, err
}

// GetNextVersionNumber 获取下一个版本号（文章维度递增）
func (r *VersionRepository) GetNextVersionNumber(articleID uint) int {
	var maxVersion int
	r.db.Model(&article.ArticleVersion{}).
		Where("article_id = ?", articleID).
		Select("COALESCE(MAX(version_number), 0)").
		Scan(&maxVersion)
	return maxVersion + 1
}

// GetVersions 获取文章的所有版本
func (r *VersionRepository) GetVersions(articleID uint) ([]article.ArticleVersion, error) {
	var versions []article.ArticleVersion
	err := r.db.Where("article_id = ?", articleID).
		Order("version_number DESC").
		Find(&versions).Error
	return versions, err
}

// UpdateStatus 更新版本状态
func (r *VersionRepository) UpdateStatus(versionID uint, status string) error {
	return r.db.Model(&article.ArticleVersion{}).
		Where("id = ?", versionID).
		Update("status", status).Error
}

// Update 更新版本完整信息
func (r *VersionRepository) Update(version *article.ArticleVersion) error {
	return r.db.Save(version).Error
}

// GetVersionNumber 获取版本的版本号
func (r *VersionRepository) GetVersionNumber(versionID uint) (int, error) {
	var version article.ArticleVersion
	err := r.db.Select("version_number").First(&version, versionID).Error
	return version.VersionNumber, err
}

// SubmissionRepository 提交审核仓储层
type SubmissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) Create(submission *article.ReviewSubmission) error {
	return r.db.Create(submission).Error
}

func (r *SubmissionRepository) GetByID(id uint) (*article.ReviewSubmission, error) {
	var submission article.ReviewSubmission
	err := r.db.First(&submission, id).Error
	return &submission, err
}

func (r *SubmissionRepository) Update(submission *article.ReviewSubmission) error {
	return r.db.Save(submission).Error
}

// ListByArticle 获取文章的所有提交（包含全部状态）
func (r *SubmissionRepository) ListByArticle(articleID uint) ([]article.ReviewSubmission, error) {
	var submissions []article.ReviewSubmission
	err := r.db.Where("article_id = ?", articleID).
		Order("created_at DESC").
		Find(&submissions).Error
	return submissions, err
}

// GetPendingByArticle 获取文章的待审核提交
func (r *SubmissionRepository) GetPendingByArticle(articleID uint) ([]article.ReviewSubmission, error) {
	var submissions []article.ReviewSubmission
	err := r.db.Where("article_id = ? AND status = ?", articleID, "pending").
		Order("created_at ASC").
		Find(&submissions).Error
	return submissions, err
}

// GetReviews 获取审核列表
func (r *SubmissionRepository) GetReviews(status string, articleID *uint) ([]article.ReviewSubmission, error) {
	query := r.db.Model(&article.ReviewSubmission{})

	if status != "all" {
		query = query.Where("status = ?", status)
	}

	if articleID != nil {
		query = query.Where("article_id = ?", *articleID)
	}

	var submissions []article.ReviewSubmission
	err := query.Order("created_at DESC").Find(&submissions).Error
	return submissions, err
}

// CreateConflict 创建冲突记录
func (r *SubmissionRepository) CreateConflict(conflict *article.VersionConflict) error {
	return r.db.Create(conflict).Error
}

// ResolveConflict 标记冲突为已解决
func (r *SubmissionRepository) ResolveConflict(submissionID uint, resolvedVersionID uint, resolvedBy uint) error {
	now := time.Now()
	return r.db.Model(&article.VersionConflict{}).
		Where("submission_id = ?", submissionID).
		Updates(map[string]interface{}{
			"status":              "resolved",
			"resolved_version_id": resolvedVersionID,
			"resolved_by":         resolvedBy,
			"resolved_at":         now,
		}).Error
}

// GetConflictBySubmission 获取提交的冲突记录
func (r *SubmissionRepository) GetConflictBySubmission(submissionID uint) (*article.VersionConflict, error) {
	var conflict article.VersionConflict
	err := r.db.Where("submission_id = ?", submissionID).First(&conflict).Error
	return &conflict, err
}

// TagRepository 标签仓储层
type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

// FindOrCreateTag 查找或创建标签
func (r *TagRepository) FindOrCreateTag(name string) (*article.Tag, error) {
	var tag article.Tag

	// 先尝试查找
	err := r.db.Where("name = ?", name).First(&tag).Error
	if err == nil {
		return &tag, nil
	}

	// 如果不存在，创建新标签
	if err == gorm.ErrRecordNotFound {
		tag = article.Tag{
			Name:      name,
			Color:     "#3b82f6", // 默认蓝色
			CreatedAt: time.Now(),
		}
		if err := r.db.Create(&tag).Error; err != nil {
			return nil, err
		}
		return &tag, nil
	}

	return nil, err
}

// AddArticleTag 添加文章标签关联
func (r *TagRepository) AddArticleTag(articleID uint, tagID uint) error {
	articleTag := &article.ArticleTag{
		ArticleID: articleID,
		TagID:     tagID,
		CreatedAt: time.Now(),
	}
	return r.db.Create(articleTag).Error
}

// RemoveArticleTags 移除文章的所有标签
func (r *TagRepository) RemoveArticleTags(articleID uint) error {
	return r.db.Where("article_id = ?", articleID).Delete(&article.ArticleTag{}).Error
}

// GetArticleTags 获取文章的所有标签
func (r *TagRepository) GetArticleTags(articleID uint) ([]article.Tag, error) {
	var tags []article.Tag
	err := r.db.
		Joins("JOIN article_tags ON article_tags.tag_id = tags.id").
		Where("article_tags.article_id = ?", articleID).
		Find(&tags).Error
	return tags, err
}


// DeleteArticleWithCascade 级联删除文章及其所有关联数据（软删除）
// 删除顺序：favorites -> article_tags -> article_collaborators -> version_conflicts -> review_submissions -> article_versions -> article
// 注意：Article 使用软删除（设置 DeletedAt），其他关联表使用硬删除
func (r *ArticleRepository) DeleteArticleWithCascade(articleID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 删除收藏记录（硬删除）
		if err := tx.Where("article_id = ?", articleID).Delete(&article.Favorite{}).Error; err != nil {
			return err
		}

		// 2. 删除文章标签关联（硬删除）
		if err := tx.Where("article_id = ?", articleID).Delete(&article.ArticleTag{}).Error; err != nil {
			return err
		}

		// 3. 删除协作者（硬删除）
		if err := tx.Where("article_id = ?", articleID).Delete(&article.ArticleCollaborator{}).Error; err != nil {
			return err
		}

		// 4. 获取所有 submission IDs 用于删除冲突记录
		var submissionIDs []uint
		if err := tx.Model(&article.ReviewSubmission{}).
			Where("article_id = ?", articleID).
			Pluck("id", &submissionIDs).Error; err != nil {
			return err
		}

		// 5. 删除版本冲突记录（硬删除）
		if len(submissionIDs) > 0 {
			if err := tx.Where("submission_id IN ?", submissionIDs).Delete(&article.VersionConflict{}).Error; err != nil {
				return err
			}
		}

		// 6. 删除审核提交记录（硬删除）
		if err := tx.Where("article_id = ?", articleID).Delete(&article.ReviewSubmission{}).Error; err != nil {
			return err
		}

		// 7. 删除所有版本（硬删除）
		if err := tx.Where("article_id = ?", articleID).Delete(&article.ArticleVersion{}).Error; err != nil {
			return err
		}

		// 8. 软删除文章本身（GORM 会自动检测 DeletedAt 字段并执行软删除）
		// 使用 Delete 方法，GORM 会自动执行 UPDATE articles SET deleted_at = NOW() WHERE id = ?
		if err := tx.Delete(&article.Article{}, articleID).Error; err != nil {
			return err
		}

		return nil
	})
}
