package article

import (
	"errors"
	"time"

	"terminal-terrace/sse-wiki/internal/dto"
	"terminal-terrace/sse-wiki/internal/model/article"
)

type ArticleService struct {
	articleRepo    *ArticleRepository
	versionRepo    *VersionRepository
	submissionRepo *SubmissionRepository
	tagRepo        *TagRepository
	mergeService   *MergeService
}

func NewArticleService(
	articleRepo *ArticleRepository,
	versionRepo *VersionRepository,
	submissionRepo *SubmissionRepository,
	tagRepo *TagRepository,
	mergeService *MergeService,
) *ArticleService {
	return &ArticleService{
		articleRepo:    articleRepo,
		versionRepo:    versionRepo,
		submissionRepo: submissionRepo,
		tagRepo:        tagRepo,
		mergeService:   mergeService,
	}
}

// CreateArticle 创建文章
func (s *ArticleService) CreateArticle(req dto.CreateArticleRequest, userID uint) (map[string]interface{}, error) {
	// 1. 创建文章记录
	art := &article.Article{
		Title:            req.Title,
		ModuleID:         req.ModuleID,
		CreatedBy:        userID,
		IsReviewRequired: req.IsReviewRequired,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.articleRepo.Create(art); err != nil {
		return nil, err
	}

	// 2. 创建初始版本
	initialVersion := &article.ArticleVersion{
		ArticleID:     art.ID,
		VersionNumber: 1,
		Content:       req.Content,
		CommitMessage: req.CommitMessage,
		AuthorID:      userID,
		Status:        "published",
		CreatedAt:     time.Now(),
	}

	if err := s.versionRepo.Create(initialVersion); err != nil {
		return nil, err
	}

	// 3. 更新文章的current_version_id
	art.CurrentVersionID = &initialVersion.ID
	if err := s.articleRepo.Update(art); err != nil {
		return nil, err
	}

	// 4. 添加创建者为owner
	if err := s.articleRepo.AddCollaborator(art.ID, userID, "owner"); err != nil {
		return nil, err
	}

	// 5. 处理标签
	if len(req.Tags) > 0 {
		for _, tagName := range req.Tags {
			if tagName == "" {
				continue
			}
			tag, err := s.tagRepo.FindOrCreateTag(tagName)
			if err != nil {
				continue
			}
			s.tagRepo.AddArticleTag(art.ID, tag.ID)
		}
	}

	// 6. 返回完整的文章详情（包括content、tags等）
	return s.GetArticle(art.ID, userID)
}

// CreateSubmission 创建提交
// 权限说明：
// - 所有用户（包括普通用户）都可以提交修改
// - 如果文章开启审核(is_review_required=true)，需要 moderator/owner 审批
// - 如果文章关闭审核(is_review_required=false)，直接发布（无需创建submission）
func (s *ArticleService) CreateSubmission(articleID uint, req dto.SubmissionRequest, userID uint) (*article.ReviewSubmission, error) {
	// 1. 获取文章信息
	art, err := s.articleRepo.GetByID(articleID)
	if err != nil {
		return nil, err
	}

	// 2. 处理标签（无论是否需要审核都先处理）
	if len(req.Tags) > 0 {
		// 移除旧标签
		s.tagRepo.RemoveArticleTags(articleID)
		// 添加新标签
		for _, tagName := range req.Tags {
			if tagName == "" {
				continue
			}
			tag, err := s.tagRepo.FindOrCreateTag(tagName)
			if err != nil {
				continue
			}
			s.tagRepo.AddArticleTag(articleID, tag.ID)
		}
	}

	// 3. 如果文章设置为免审核，直接发布新版本（无需创建submission）
	if !art.IsReviewRequired {
		// 创建published版本
		publishedVersion := &article.ArticleVersion{
			ArticleID:     articleID,
			VersionNumber: s.versionRepo.GetNextVersionNumber(articleID),
			Content:       req.Content,
			CommitMessage: req.CommitMessage,
			AuthorID:      userID,
			Status:        "published",
			CreatedAt:     time.Now(),
		}

		if err := s.versionRepo.Create(publishedVersion); err != nil {
			return nil, err
		}

		// 更新文章的current_version_id
		art.CurrentVersionID = &publishedVersion.ID
		art.UpdatedAt = time.Now()
		if err := s.articleRepo.Update(art); err != nil {
			return nil, err
		}

		// 返回nil表示无需审核（直接发布成功）
		return nil, nil
	}

	// 4. 需要审核：创建draft版本和审核提交记录
	draftVersion := &article.ArticleVersion{
		ArticleID:     articleID,
		VersionNumber: s.versionRepo.GetNextVersionNumber(articleID),
		Content:       req.Content,
		CommitMessage: req.CommitMessage,
		AuthorID:      userID,
		Status:        "draft",
		CreatedAt:     time.Now(),
	}

	if err := s.versionRepo.Create(draftVersion); err != nil {
		return nil, err
	}

	// 5. 创建审核提交记录
	submission := &article.ReviewSubmission{
		ArticleID:         articleID,
		ProposedVersionID: draftVersion.ID,
		BaseVersionID:     req.BaseVersionID,
		SubmittedBy:       userID,
		Status:            "pending",
		CreatedAt:         time.Now(),
	}

	if err := s.submissionRepo.Create(submission); err != nil {
		return nil, err
	}

	return submission, nil
}

// ReviewSubmission 审核提交
func (s *ArticleService) ReviewSubmission(submissionID uint, reviewerID uint, userRole string, req dto.ReviewActionRequest) (interface{}, error) {
	// 1. 获取submission
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return nil, err
	}

	// 2. 获取文章信息
	art, err := s.articleRepo.GetByID(submission.ArticleID)
	if err != nil {
		return nil, err
	}

	// 3. 检查权限（全局admin或文章moderator及以上）
	// 注意：免审核模式下的自动合并不需要权限检查（提交者即审核者）
	isAutoApprove := !art.IsReviewRequired && submission.SubmittedBy == reviewerID
	if !isAutoApprove {
		hasPermission := s.articleRepo.CheckPermission(submission.ArticleID, reviewerID, userRole, "moderator")
		if !hasPermission {
			return nil, errors.New("permission denied")
		}
	}

	if req.Action == "approve" {
		// === 执行3路合并 ===

		// 获取三方内容
		baseContent, err := s.versionRepo.GetContent(submission.BaseVersionID)
		if err != nil {
			return nil, err
		}

		theirContent, err := s.versionRepo.GetContent(submission.ProposedVersionID)
		if err != nil {
			return nil, err
		}

		ourContent, err := s.versionRepo.GetContent(*art.CurrentVersionID)
		if err != nil {
			return nil, err
		}

		// 调用3路合并算法
		mergeResult := s.mergeService.ThreeWayMerge(baseContent, theirContent, ourContent)

		if mergeResult.HasConflict && req.MergedContent == nil {
			// === 有冲突且管理员未提供解决方案：返回冲突信息 ===

			// 更新submission状态
			submission.Status = "conflict_detected"
			submission.HasConflict = true
			submission.MergeResult = mergeResult.ConflictMarkedContent
			s.submissionRepo.Update(submission)

			// 记录冲突
			conflict := &article.VersionConflict{
				SubmissionID:          submission.ID,
				ConflictWithVersionID: *art.CurrentVersionID,
				Status:                "detected",
				ConflictDetails:       mergeResult.ConflictDetails,
				CreatedAt:             time.Now(),
			}
			s.submissionRepo.CreateConflict(conflict)

			// 获取版本号信息
			baseVersionNumber, _ := s.versionRepo.GetVersionNumber(submission.BaseVersionID)
			currentVersionNumber, _ := s.versionRepo.GetVersionNumber(*art.CurrentVersionID)

			// 返回冲突错误
			return nil, &MergeConflictError{
				Message: "Merge conflict detected",
				ConflictData: map[string]interface{}{
					"base_content":            baseContent,
					"their_content":           theirContent,
					"our_content":             ourContent,
					"merged_content":          mergeResult.ConflictMarkedContent,
					"has_conflict":            true,
					"base_version_number":     baseVersionNumber,
					"current_version_number":  currentVersionNumber,
					"submitter_name":          "User",
				},
			}
		}

		// === 无冲突或管理员已解决冲突：执行合并 ===

		var finalContent string
		if req.MergedContent != nil && *req.MergedContent != "" {
			// 管理员手动解决了冲突
			finalContent = *req.MergedContent
		} else {
			// 自动合并成功
			finalContent = mergeResult.MergedContent
		}

		// 获取提交的draft版本
		proposedVersion, err := s.versionRepo.GetByID(submission.ProposedVersionID)
		if err != nil {
			return nil, err
		}

		// 直接将draft版本改为published状态，更新content（如果有冲突解决）
		proposedVersion.Status = "published"
		proposedVersion.Content = finalContent

		if err := s.versionRepo.Update(proposedVersion); err != nil {
			return nil, err
		}

		// 更新文章的current_version_id指向这个版本
		art.CurrentVersionID = &proposedVersion.ID
		art.UpdatedAt = time.Now()
		if err := s.articleRepo.Update(art); err != nil {
			return nil, err
		}

		// 更新submission状态
		submission.Status = "merged"
		submission.ReviewedBy = &reviewerID
		now := time.Now()
		submission.ReviewedAt = &now
		submission.ReviewNotes = req.Notes
		s.submissionRepo.Update(submission)

		// 如果之前有冲突记录，标记为已解决
		if submission.HasConflict {
			s.submissionRepo.ResolveConflict(submission.ID, proposedVersion.ID, reviewerID)
		}

		return map[string]interface{}{
			"message":            "Successfully merged and published",
			"published_version_id": proposedVersion.ID,
		}, nil

	} else if req.Action == "reject" {
		// 驳回提交
		submission.Status = "rejected"
		submission.ReviewedBy = &reviewerID
		now := time.Now()
		submission.ReviewedAt = &now
		submission.ReviewNotes = req.Notes
		s.submissionRepo.Update(submission)

		// 更新版本状态为rejected（该版本不会被发布，不更新current_version_id）
		s.versionRepo.UpdateStatus(submission.ProposedVersionID, "rejected")

		return map[string]interface{}{"message": "Submission rejected"}, nil
	}

	return nil, errors.New("invalid action")
}

// autoApprove 自动审核（免审核模式）
func (s *ArticleService) autoApprove(submissionID uint, userID uint, userRole string) (*article.ReviewSubmission, error) {
	_, err := s.ReviewSubmission(submissionID, userID, userRole, dto.ReviewActionRequest{
		Action: "approve",
	})

	if err != nil {
		// 如果是冲突错误，需要用户处理
		if _, ok := err.(*MergeConflictError); ok {
			return nil, err
		}
		return nil, err
	}

	submission, _ := s.submissionRepo.GetByID(submissionID)
	return submission, nil
}

// GetArticle 获取文章详情
func (s *ArticleService) GetArticle(articleID uint, userID uint) (map[string]interface{}, error) {
	art, err := s.articleRepo.GetByID(articleID)
	if err != nil {
		return nil, err
	}

	// 获取用户在该文章的角色
	role := s.articleRepo.GetUserRole(articleID, userID)

	// 获取待审核的提交，并为每个提交加载 proposed_version 信息
	pendingSubmissions, _ := s.submissionRepo.GetPendingByArticle(articleID)

	// 为每个 submission 加载 proposed_version 的完整信息
	submissionsWithVersion := make([]map[string]interface{}, len(pendingSubmissions))
	for i, submission := range pendingSubmissions {
		proposedVersion, err := s.versionRepo.GetByID(submission.ProposedVersionID)
		submissionMap := map[string]interface{}{
			"id":                  submission.ID,
			"article_id":          submission.ArticleID,
			"proposed_version_id": submission.ProposedVersionID,
			"base_version_id":     submission.BaseVersionID,
			"submitted_by":        submission.SubmittedBy,
			"reviewed_by":         submission.ReviewedBy,
			"status":              submission.Status,
			"review_notes":        submission.ReviewNotes,
			"has_conflict":        submission.HasConflict,
			"created_at":          submission.CreatedAt,
			"reviewed_at":         submission.ReviewedAt,
		}

		// 添加 proposed_version 完整信息（包括 content）
		if err == nil {
			submissionMap["proposed_version"] = map[string]interface{}{
				"id":             proposedVersion.ID,
				"article_id":     proposedVersion.ArticleID,
				"version_number": proposedVersion.VersionNumber,
				"content":        proposedVersion.Content,
				"commit_message": proposedVersion.CommitMessage,
				"author_id":      proposedVersion.AuthorID,
				"status":         proposedVersion.Status,
				"created_at":     proposedVersion.CreatedAt,
			}
		}

		submissionsWithVersion[i] = submissionMap
	}

	// 获取文章标签
	tags, _ := s.tagRepo.GetArticleTags(articleID)
	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	// 获取当前版本的内容
	var content string
	var commitMessage string
	var versionNumber int
	if art.CurrentVersionID != nil {
		currentVersion, err := s.versionRepo.GetByID(*art.CurrentVersionID)
		if err == nil {
			content = currentVersion.Content
			commitMessage = currentVersion.CommitMessage
			versionNumber = currentVersion.VersionNumber
		}
	}

	// 增加阅读量
	s.articleRepo.IncrementViewCount(articleID)

	roleStr := ""
	if role != nil {
		roleStr = *role
	}

	return map[string]interface{}{
		"id":                  art.ID,
		"title":               art.Title,
		"module_id":           art.ModuleID,
		"content":             content,
		"commit_message":      commitMessage,
		"version_number":      versionNumber,
		"current_version_id":  art.CurrentVersionID,
		"current_user_role":   roleStr,
		"is_review_required":  art.IsReviewRequired,
		"view_count":          art.ViewCount,
		"tags":                tagNames,
		"pending_submissions": submissionsWithVersion,
		"created_by":          art.CreatedBy,
		"created_at":          art.CreatedAt,
		"updated_at":          art.UpdatedAt,
	}, nil
}

// GetArticlesByModule 获取模块下的文章列表（分页）
func (s *ArticleService) GetArticlesByModule(moduleID uint, page, pageSize int) (map[string]interface{}, error) {
	offset := (page - 1) * pageSize

	articles, total, err := s.articleRepo.ListByModuleID(moduleID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	articleItems := make([]map[string]interface{}, len(articles))
	for i, art := range articles {
		// 获取每个文章的标签
		tags, _ := s.tagRepo.GetArticleTags(art.ID)
		tagNames := make([]string, len(tags))
		for j, tag := range tags {
			tagNames[j] = tag.Name
		}

		articleItems[i] = map[string]interface{}{
			"id":                 art.ID,
			"title":              art.Title,
			"module_id":          art.ModuleID,
			"current_version_id": art.CurrentVersionID,
			"view_count":         art.ViewCount,
			"created_by":         art.CreatedBy,
			"tags":               tagNames,
			"created_at":         art.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":         art.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return map[string]interface{}{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"articles":  articleItems,
	}, nil
}

// GetReviews 获取审核列表
func (s *ArticleService) GetReviews(status string, articleID *uint) ([]article.ReviewSubmission, error) {
	return s.submissionRepo.GetReviews(status, articleID)
}

// GetReviewDetail 获取审核详情（包含proposed_version完整信息）
func (s *ArticleService) GetReviewDetail(submissionID uint) (map[string]interface{}, error) {
	// 1. 获取 submission
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return nil, err
	}

	// 2. 获取 proposed_version 完整信息
	proposedVersion, err := s.versionRepo.GetByID(submission.ProposedVersionID)
	if err != nil {
		return nil, err
	}

	// 3. 构造返回数据
	result := map[string]interface{}{
		"id":                  submission.ID,
		"article_id":          submission.ArticleID,
		"proposed_version_id": submission.ProposedVersionID,
		"base_version_id":     submission.BaseVersionID,
		"submitted_by":        submission.SubmittedBy,
		"reviewed_by":         submission.ReviewedBy,
		"status":              submission.Status,
		"review_notes":        submission.ReviewNotes,
		"has_conflict":        submission.HasConflict,
		"merge_result":        submission.MergeResult,
		"created_at":          submission.CreatedAt,
		"reviewed_at":         submission.ReviewedAt,
		"proposed_version": map[string]interface{}{
			"id":             proposedVersion.ID,
			"article_id":     proposedVersion.ArticleID,
			"version_number": proposedVersion.VersionNumber,
			"content":        proposedVersion.Content,
			"commit_message": proposedVersion.CommitMessage,
			"author_id":      proposedVersion.AuthorID,
			"status":         proposedVersion.Status,
			"created_at":     proposedVersion.CreatedAt,
		},
	}

	return result, nil
}

// UpdateSettings 更新文章设置
func (s *ArticleService) UpdateSettings(articleID uint, userID uint, userRole string, req dto.UpdateArticleSettingsRequest) error {
	// 检查权限（全局admin或文章owner）
	hasPermission := s.articleRepo.CheckPermission(articleID, userID, userRole, "owner")
	if !hasPermission {
		return errors.New("permission denied")
	}

	art, err := s.articleRepo.GetByID(articleID)
	if err != nil {
		return err
	}

	if req.IsReviewRequired != nil {
		art.IsReviewRequired = *req.IsReviewRequired
	}

	return s.articleRepo.Update(art)
}

// AddCollaborator 添加协作者
func (s *ArticleService) AddCollaborator(articleID uint, userID uint, userRole string, req dto.AddCollaboratorRequest) error {
	// 检查权限（全局admin或文章owner）
	hasPermission := s.articleRepo.CheckPermission(articleID, userID, userRole, "owner")
	if !hasPermission {
		return errors.New("permission denied")
	}

	return s.articleRepo.AddCollaborator(articleID, req.UserID, req.Role)
}

// GetVersions 获取文章版本列表
func (s *ArticleService) GetVersions(articleID uint) ([]article.ArticleVersion, error) {
	return s.versionRepo.GetVersions(articleID)
}

// GetVersionByID 获取特定版本
func (s *ArticleService) GetVersionByID(versionID uint) (*article.ArticleVersion, error) {
	return s.versionRepo.GetByID(versionID)
}

// MergeConflictError 自定义冲突错误
type MergeConflictError struct {
	Message      string
	ConflictData map[string]interface{}
}

func (e *MergeConflictError) Error() string {
	return e.Message
}
