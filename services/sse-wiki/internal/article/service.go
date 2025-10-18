package article

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
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

	// 2. 创建初始版本（v1，base_version_id为nil）
	initialVersion := &article.ArticleVersion{
		ArticleID:     art.ID,
		VersionNumber: 1,
		Content:       req.Content,
		CommitMessage: req.CommitMessage,
		AuthorID:      userID,
		Status:        "published",
		BaseVersionID: nil, // v1没有基础版本
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
	// 创建者自动成为 owner，传空字符串让 GetArticle 从数据库读取角色
	return s.GetArticle(art.ID, userID, "")
}

// CreateSubmission 创建提交
// 权限说明：
// - 所有用户（包括普通用户）都可以提交修改
// - 如果提交者是 admin/owner/moderator，直接发布
// - 如果文章开启审核(is_review_required=true) 且提交者是普通用户，需要审核
// - 如果文章关闭审核(is_review_required=false)，直接发布（执行3路合并）
func (s *ArticleService) CreateSubmission(articleID uint, req dto.SubmissionRequest, userID uint, userRole string) (*article.ReviewSubmission, *article.ArticleVersion, error) {
	// 1. 获取文章信息
	art, err := s.articleRepo.GetByID(articleID)
	if err != nil {
		return nil, nil, err
	}

	// 2. 检查提交者权限
	userArticleRole := s.articleRepo.GetUserRole(articleID, userID)
	var userArticleRoleStr string
	if userArticleRole != nil {
		userArticleRoleStr = *userArticleRole
	}

	// 判断是否需要审核
	// - 全局admin：直接发布
	// - owner/moderator：直接发布
	// - 文章设置免审核：直接发布
	// - 其他情况：需要审核
	isAdmin := userRole == "admin"
	isOwnerOrModerator := userArticleRoleStr == "owner" || userArticleRoleStr == "moderator"
	isReviewRequired := art.IsReviewRequired != nil && *art.IsReviewRequired
	needReview := isReviewRequired && !isAdmin && !isOwnerOrModerator

	// TODO: 生产环境优化 - 替换为结构化日志库（如zap/zerolog），并根据环境变量控制日志级别
	log.Printf("[CreateSubmission] articleID=%d, userID=%d, userRole=%s, articleRole=%s, isReviewRequired=%v, needReview=%v",
		articleID, userID, userRole, userArticleRoleStr, isReviewRequired, needReview)

	// 3. 如果不需要审核（免审核模式 或 管理员/owner/moderator），执行3路合并并直接发布
	if !needReview {
		// TODO: 生产环境优化 - 移除或使用结构化日志
		log.Printf("[CreateSubmission] 直接发布模式: userRole=%s, articleRole=%s", userRole, userArticleRoleStr)
		// 获取三方内容
		baseContent, err := s.versionRepo.GetContent(req.BaseVersionID)
		if err != nil {
			return nil, nil, errors.New("无效的基础版本ID")
		}

		theirContent := req.Content

		ourContent := ""
		if art.CurrentVersionID != nil {
			ourContent, err = s.versionRepo.GetContent(*art.CurrentVersionID)
			if err != nil {
				return nil, nil, errors.New("无法获取当前版本内容")
			}
		}

		// 执行3路合并
		mergeResult := s.mergeService.ThreeWayMerge(baseContent, theirContent, ourContent)

		if mergeResult.HasConflict {
			// TODO: 生产环境优化 - 移除或使用结构化日志
			log.Printf("[CreateSubmission] 3路合并检测到冲突, articleID=%d", articleID)

			baseVersionNumber, _ := s.versionRepo.GetVersionNumber(req.BaseVersionID)
			currentVersionNumber := 0
			if art.CurrentVersionID != nil {
				if num, err := s.versionRepo.GetVersionNumber(*art.CurrentVersionID); err == nil {
					currentVersionNumber = num
				}
			}

			return nil, nil, &MergeConflictError{
				Message: "Merge conflict detected",
				ConflictData: map[string]interface{}{
					"base_content":           baseContent,
					"their_content":          theirContent,
					"our_content":            ourContent,
					"has_conflict":           true,
					"base_version_number":    baseVersionNumber,
					"current_version_number": currentVersionNumber,
					"submitter_name":         "User",
				},
			}
		}

		// 创建published版本（使用合并后的内容）
		nextVersionNumber := s.versionRepo.GetNextVersionNumber(articleID)
		publishedVersion := &article.ArticleVersion{
			ArticleID:              articleID,
			VersionNumber:          nextVersionNumber,
			Content:                mergeResult.MergedContent,
			CommitMessage:          req.CommitMessage,
			AuthorID:               userID,
			Status:                 "published",
			BaseVersionID:          &req.BaseVersionID,
			MergedAgainstVersionID: art.CurrentVersionID,
			CreatedAt:              time.Now(),
		}

		if err := s.versionRepo.Create(publishedVersion); err != nil {
			// 检查是否是版本号冲突
			if isVersionConflictError(err) {
				// TODO: 生产环境优化 - 移除或使用结构化日志
				log.Printf("[CreateSubmission] 版本号冲突, articleID=%d, versionNumber=%d", articleID, nextVersionNumber)
				return nil, nil, errors.New("版本号冲突，请刷新后重试")
			}
			return nil, nil, err
		}

		// TODO: 生产环境优化 - 移除或使用结构化日志
		log.Printf("[CreateSubmission] 创建published版本成功, versionID=%d, versionNumber=%d", publishedVersion.ID, publishedVersion.VersionNumber)

		// 更新文章的current_version_id
		art.CurrentVersionID = &publishedVersion.ID
		art.UpdatedAt = time.Now()
		if err := s.articleRepo.Update(art); err != nil {
			return nil, nil, err
		}

		// 免审核模式：立即应用标签
		if len(req.Tags) > 0 {
			s.tagRepo.RemoveArticleTags(articleID)
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

		// 返回nil表示无需审核（直接发布成功）
		return nil, publishedVersion, nil
	}

	// 4. 需要审核：创建pending版本和审核提交记录
	// TODO: 生产环境优化 - 移除或使用结构化日志
	log.Printf("[CreateSubmission] 需要审核模式: 创建pending版本")
	nextVersionNumber := s.versionRepo.GetNextVersionNumber(articleID)
	pendingVersion := &article.ArticleVersion{
		ArticleID:     articleID,
		VersionNumber: nextVersionNumber,
		Content:       req.Content,
		CommitMessage: req.CommitMessage,
		AuthorID:      userID,
		Status:        "pending",
		BaseVersionID: &req.BaseVersionID, // 记录基于哪个版本创建
		CreatedAt:     time.Now(),
	}

	if err := s.versionRepo.Create(pendingVersion); err != nil {
		// 检查是否是版本号冲突
		if isVersionConflictError(err) {
			// TODO: 生产环境优化 - 移除或使用结构化日志
			log.Printf("[CreateSubmission] 版本号冲突, articleID=%d, versionNumber=%d", articleID, nextVersionNumber)
			return nil, nil, errors.New("版本号冲突，请刷新后重试")
		}
		return nil, nil, err
	}

	// TODO: 生产环境优化 - 移除或使用结构化日志
	log.Printf("[CreateSubmission] 创建pending版本成功, versionID=%d, versionNumber=%d", pendingVersion.ID, pendingVersion.VersionNumber)

	// 5. 序列化标签为JSON（审核通过后才应用）
	tagsJSON := "[]"
	if len(req.Tags) > 0 {
		tagsBytes, err := json.Marshal(req.Tags)
		if err == nil {
			tagsJSON = string(tagsBytes)
		}
	}

	// 6. 创建审核提交记录
	submission := &article.ReviewSubmission{
		ArticleID:         articleID,
		ProposedVersionID: pendingVersion.ID,
		BaseVersionID:     req.BaseVersionID,
		ProposedTags:      tagsJSON,
		SubmittedBy:       userID,
		Status:            "pending",
		CreatedAt:         time.Now(),
	}

	if err := s.submissionRepo.Create(submission); err != nil {
		return nil, nil, err
	}

	// TODO: 生产环境优化 - 移除或使用结构化日志
	log.Printf("[CreateSubmission] 创建审核提交成功, submissionID=%d, 等待审核", submission.ID)

	return submission, nil, nil
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
	isReviewRequired := art.IsReviewRequired != nil && *art.IsReviewRequired
	isAutoApprove := !isReviewRequired && submission.SubmittedBy == reviewerID
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
			submission.MergeResult = ""
			submission.MergedAgainstVersionID = art.CurrentVersionID
			s.submissionRepo.Update(submission)

			// 记录冲突
			conflict := &article.VersionConflict{
				SubmissionID:          submission.ID,
				ConflictWithVersionID: *art.CurrentVersionID,
				Status:                "detected",
				ConflictDetails:       "",
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
					"base_content":           baseContent,
					"their_content":          theirContent,
					"our_content":            ourContent,
					"has_conflict":           true,
					"base_version_number":    baseVersionNumber,
					"current_version_number": currentVersionNumber,
					// TODO: 用户服务集成 - 从用户服务获取真实的用户名而不是硬编码
					"submitter_name": "User",
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

		// 获取提交的pending版本
		proposedVersion, err := s.versionRepo.GetByID(submission.ProposedVersionID)
		if err != nil {
			return nil, err
		}

		// 直接将pending版本改为published状态，更新content（如果有冲突解决）
		proposedVersion.Status = "published"
		proposedVersion.Content = finalContent
		proposedVersion.MergedAgainstVersionID = art.CurrentVersionID

		if err := s.versionRepo.Update(proposedVersion); err != nil {
			return nil, err
		}

		// 更新文章的current_version_id指向这个版本
		art.CurrentVersionID = &proposedVersion.ID
		art.UpdatedAt = time.Now()
		if err := s.articleRepo.Update(art); err != nil {
			return nil, err
		}

		// 审核通过后应用标签
		var proposedTags []string
		if submission.ProposedTags != "" && submission.ProposedTags != "[]" {
			if err := json.Unmarshal([]byte(submission.ProposedTags), &proposedTags); err == nil {
				// 移除旧标签
				s.tagRepo.RemoveArticleTags(submission.ArticleID)
				// 添加新标签
				for _, tagName := range proposedTags {
					if tagName == "" {
						continue
					}
					tag, err := s.tagRepo.FindOrCreateTag(tagName)
					if err != nil {
						continue
					}
					s.tagRepo.AddArticleTag(submission.ArticleID, tag.ID)
				}
			}
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
			"message":           "Successfully merged and published",
			"published_version": proposedVersion,
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

// GetArticle 获取文章详情
func (s *ArticleService) GetArticle(articleID uint, userID uint, globalUserRole string) (map[string]interface{}, error) {
	art, err := s.articleRepo.GetByID(articleID)
	if err != nil {
		return nil, err
	}

	// 获取用户在该文章的角色
	articleRole := s.articleRepo.GetUserRole(articleID, userID)

	// 综合判断用户角色：全局 admin 或文章 owner/moderator
	effectiveRole := ""
	if globalUserRole == "admin" {
		effectiveRole = "admin"
	} else if articleRole != nil {
		effectiveRole = *articleRole
	}

	// 获取所有提交
	allSubmissions, _ := s.submissionRepo.ListByArticle(articleID)

	// 获取所有版本
	allVersions, _ := s.versionRepo.GetVersions(articleID)

	// 构造历史记录（版本 + 提交）
	historyEntries := make([]map[string]interface{}, 0, len(allVersions)+len(allSubmissions))

	// 只添加 published 和 rejected 状态的版本到历史列表
	// pending 状态的版本属于 submission 的 proposed_version，通过 submission 条目展示
	for _, v := range allVersions {
		// 跳过 pending 状态的版本，避免与 submission 重复显示
		if v.Status == "pending" {
			continue
		}

		historyEntries = append(historyEntries, map[string]interface{}{
			"entry_type":                "version",
			"entry_id":                  v.ID,
			"version_id":                v.ID,
			"submission_id":             nil,
			"status":                    v.Status,
			"submission_status":         nil,
			"base_version_id":           v.BaseVersionID,
			"merged_against_version_id": v.MergedAgainstVersionID,
			"has_conflict":              false,
			"merge_result":              nil,
			"commit_message":            v.CommitMessage,
			"author_id":                 v.AuthorID,
			"created_at":                v.CreatedAt,
		})
	}

	// 只添加未合并的 submission 到历史列表
	// merged 状态的 submission 已经作为 published version 显示，不需要重复
	for _, submission := range allSubmissions {
		// 跳过已合并的 submission，避免与 published version 重复显示
		if submission.Status == "merged" {
			continue
		}

		entry := map[string]interface{}{
			"entry_type":                "submission",
			"entry_id":                  submission.ID,
			"version_id":                submission.ProposedVersionID,
			"submission_id":             submission.ID,
			"status":                    nil,
			"submission_status":         submission.Status,
			"base_version_id":           submission.BaseVersionID,
			"merged_against_version_id": submission.MergedAgainstVersionID,
			"has_conflict":              submission.HasConflict,
			"merge_result":              submission.MergeResult,
			"commit_message":            "",
			"author_id":                 submission.SubmittedBy,
			"reviewed_by":               submission.ReviewedBy,
			"review_notes":              submission.ReviewNotes,
			"created_at":                submission.CreatedAt,
			"reviewed_at":               submission.ReviewedAt,
		}

		// 补充提交版本信息（commit_message）
		if proposedVersion, err := s.versionRepo.GetByID(submission.ProposedVersionID); err == nil {
			entry["commit_message"] = proposedVersion.CommitMessage
		}

		historyEntries = append(historyEntries, entry)
	}

	// 按创建时间倒序排序
	sort.Slice(historyEntries, func(i, j int) bool {
		ti := historyEntries[i]["created_at"].(time.Time)
		tj := historyEntries[j]["created_at"].(time.Time)
		return ti.After(tj)
	})

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

	return map[string]interface{}{
		"id":                 art.ID,
		"title":              art.Title,
		"module_id":          art.ModuleID,
		"content":            content,
		"commit_message":     commitMessage,
		"version_number":     versionNumber,
		"current_version_id": art.CurrentVersionID,
		"current_user_role":  effectiveRole,
		"is_review_required": art.IsReviewRequired,
		"view_count":         art.ViewCount,
		"tags":               tagNames,
		"created_by":         art.CreatedBy,
		"created_at":         art.CreatedAt,
		"updated_at":         art.UpdatedAt,
		"history":            historyEntries,
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
func (s *ArticleService) GetReviewDetail(submissionID uint, userID uint, globalUserRole string) (map[string]interface{}, error) {
	// 1. 获取 submission
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		// TODO: 生产环境优化 - 移除或使用结构化日志
		log.Printf("[GetReviewDetail] 获取submission失败: submissionID=%d, error=%v", submissionID, err)
		return nil, errors.New("提交记录不存在")
	}

	// TODO: 生产环境优化 - 移除或使用结构化日志
	log.Printf("[GetReviewDetail] submissionID=%d, status=%s, proposedVersionID=%d, userID=%d",
		submissionID, submission.Status, submission.ProposedVersionID, userID)

	// 2. 获取 proposed_version 完整信息
	proposedVersion, err := s.versionRepo.GetByID(submission.ProposedVersionID)
	if err != nil {
		// TODO: 生产环境优化 - 移除或使用结构化日志
		log.Printf("[GetReviewDetail] 获取proposed_version失败: versionID=%d, error=%v",
			submission.ProposedVersionID, err)
		return nil, fmt.Errorf("提交的版本(ID=%d)不存在，数据可能已损坏。建议联系管理员检查数据库",
			submission.ProposedVersionID)
	}

	// 3. 获取 base_version
	var baseVersion *article.ArticleVersion
	if submission.BaseVersionID > 0 {
		baseVersion, err = s.versionRepo.GetByID(submission.BaseVersionID)
		if err != nil {
			// TODO: 生产环境优化 - 移除或使用结构化日志
			log.Printf("[GetReviewDetail] 警告: 获取base_version失败: versionID=%d, error=%v",
				submission.BaseVersionID, err)
			// base_version 不存在时不报错，返回 nil
			baseVersion = nil
		}
	}

	// 4. 获取文章当前版本（current_version）
	art, err := s.articleRepo.GetByID(submission.ArticleID)
	if err != nil {
		// TODO: 生产环境优化 - 移除或使用结构化日志
		log.Printf("[GetReviewDetail] 获取article失败: articleID=%d, error=%v",
			submission.ArticleID, err)
		return nil, errors.New("文章不存在")
	}

	var currentVersion *article.ArticleVersion
	if art.CurrentVersionID != nil {
		currentVersion, err = s.versionRepo.GetByID(*art.CurrentVersionID)
		if err != nil {
			// TODO: 生产环境优化 - 移除或使用结构化日志
			log.Printf("[GetReviewDetail] 警告: 获取current_version失败: versionID=%d, error=%v",
				*art.CurrentVersionID, err)
			// current_version 不存在时不报错，返回 nil
			currentVersion = nil
		}
	}

	// 5. 实时检测冲突（每次获取审核详情时重新执行三路合并）
	var realTimeHasConflict bool
	var realTimeMergeResult string

	// 只有在 pending 或 conflict_detected 状态时才需要实时检测
	if submission.Status == "pending" || submission.Status == "conflict_detected" {
		if baseVersion != nil && currentVersion != nil && proposedVersion != nil {
			// 执行三路合并检测
			baseContent := baseVersion.Content
			theirContent := proposedVersion.Content
			ourContent := currentVersion.Content

			mergeResult := s.mergeService.ThreeWayMerge(baseContent, theirContent, ourContent)
			realTimeHasConflict = mergeResult.HasConflict

			if realTimeHasConflict {
				// 冲突直接返回三方原始内容，由前端生成冲突标记
				// TODO: 生产环境优化 - 移除或使用结构化日志
				log.Printf("[GetReviewDetail] 实时检测到冲突: submissionID=%d, base=%d, current=%d, proposed=%d",
					submissionID, baseVersion.ID, currentVersion.ID, proposedVersion.ID)
			} else {
				// 无冲突返回合并后的最终内容
				realTimeMergeResult = mergeResult.MergedContent
				// TODO: 生产环境优化 - 移除或使用结构化日志
				log.Printf("[GetReviewDetail] 实时检测无冲突: submissionID=%d", submissionID)
			}
		}
	} else {
		// 已审核的使用存储的结果
		realTimeHasConflict = submission.HasConflict
		realTimeMergeResult = submission.MergeResult
	}

	// 6. 计算用户在该文章的有效角色
	articleRole := s.articleRepo.GetUserRole(submission.ArticleID, userID)
	effectiveRole := ""
	if globalUserRole == "admin" {
		effectiveRole = "admin"
	} else if articleRole != nil {
		effectiveRole = *articleRole
	}

	// 7. 构造返回数据
	result := map[string]interface{}{
		"id":                  submission.ID,
		"article_id":          submission.ArticleID,
		"proposed_version_id": submission.ProposedVersionID,
		"base_version_id":     submission.BaseVersionID,
		"submitted_by":        submission.SubmittedBy,
		"reviewed_by":         submission.ReviewedBy,
		"status":              submission.Status,
		"review_notes":        submission.ReviewNotes,
		"has_conflict":        realTimeHasConflict, // 使用实时检测结果
		"merge_result":        realTimeMergeResult,
		"created_at":          submission.CreatedAt,
		"reviewed_at":         submission.ReviewedAt,
		"proposed_version":    proposedVersion,
		"base_version":        baseVersion,
		"current_version":     currentVersion,
		"current_user_role":   effectiveRole, // 返回用户角色，前端用于权限判断
	}

	// 7. 如果有冲突，返回三方原始内容（不生成冲突标记，由前端处理）
	if realTimeHasConflict {
		// 获取三方内容用于展示冲突
		var baseContent, theirContent, ourContent string

		if baseVersion != nil {
			baseContent = baseVersion.Content
		}

		if proposedVersion != nil {
			theirContent = proposedVersion.Content
		}

		if currentVersion != nil {
			ourContent = currentVersion.Content
		}

		// 获取版本号
		var baseVersionNumber, currentVersionNumber int
		if baseVersion != nil {
			baseVersionNumber = baseVersion.VersionNumber
		}
		if currentVersion != nil {
			currentVersionNumber = currentVersion.VersionNumber
		}

		result["conflict_data"] = map[string]interface{}{
			"base_content":           baseContent,
			"their_content":          theirContent,
			"our_content":            ourContent,
			"has_conflict":           true,
			"base_version_number":    baseVersionNumber,
			"current_version_number": currentVersionNumber,
		}

		// TODO: 生产环境优化 - 移除或使用结构化日志
		log.Printf("[GetReviewDetail] 返回冲突数据: submissionID=%d, base=%d, current=%d, proposed=%d",
			submissionID, baseVersion.ID, currentVersion.ID, proposedVersion.ID)
	}

	// TODO: 生产环境优化 - 移除或使用结构化日志
	log.Printf("[GetReviewDetail] 成功返回审核详情: submissionID=%d", submissionID)
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
		art.IsReviewRequired = req.IsReviewRequired
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

// GetVersionDiff 获取版本的diff信息
func (s *ArticleService) GetVersionDiff(versionID uint) (map[string]interface{}, error) {
	// 1. 获取当前版本
	version, err := s.versionRepo.GetByID(versionID)
	if err != nil {
		return nil, errors.New("版本不存在")
	}

	// 2. 构造响应数据
	result := map[string]interface{}{
		"current_version": version,
	}

	// 3. 如果有base_version_id，获取基础版本
	if version.BaseVersionID != nil {
		baseVersion, err := s.versionRepo.GetByID(*version.BaseVersionID)
		if err != nil {
			// 如果base_version不存在，返回null（数据一致性问题）
			// TODO: 生产环境优化 - 移除或使用结构化日志
			log.Printf("警告: 版本%d的base_version_id=%d不存在", versionID, *version.BaseVersionID)
			result["base_version"] = nil
		} else {
			result["base_version"] = baseVersion
		}
	} else {
		// v1初始版本，没有base_version
		result["base_version"] = nil
	}

	return result, nil
}

// MergeConflictError 自定义冲突错误
type MergeConflictError struct {
	Message      string
	ConflictData map[string]interface{}
}

func (e *MergeConflictError) Error() string {
	return e.Message
}

// isVersionConflictError 检测是否是版本号冲突错误
func isVersionConflictError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// 检测MySQL和PostgreSQL的唯一索引冲突错误
	return containsAny(errMsg, []string{
		"Duplicate entry",
		"duplicate key value",
		"UNIQUE constraint failed",
		"idx_article_version_unique",
	})
}

// containsAny 检查字符串是否包含任意一个子串
// TODO: 代码优化 - 可以使用 strings.Contains 简化实现
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
