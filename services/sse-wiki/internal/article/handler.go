package article

import (
	"strconv"
	"terminal-terrace/response"
	"terminal-terrace/sse-wiki/internal/dto"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleHandler struct {
	articleService *ArticleService
	tagRepo        *TagRepository
}

func NewArticleHandler(db *gorm.DB) *ArticleHandler {
	articleRepo := NewArticleRepository(db)
	versionRepo := NewVersionRepository(db)
	submissionRepo := NewSubmissionRepository(db)
	tagRepo := NewTagRepository(db)
	mergeService := NewMergeService()

	return &ArticleHandler{
		articleService: NewArticleService(articleRepo, versionRepo, submissionRepo, tagRepo, mergeService),
		tagRepo:        tagRepo,
	}
}

// GetArticlesByModule 获取模块下的文章列表
// @Summary 获取模块下的文章列表（分页）
// @Tags Article
// @Accept json
// @Produce json
// @Param id path int true "模块ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=dto.ArticleListResponse}
// @Router /modules/{id}/articles [get]
func (h *ArticleHandler) GetArticlesByModule(c *gin.Context) {
	moduleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的模块ID"),
		))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.articleService.GetArticlesByModule(uint(moduleID), page, pageSize)
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取文章列表失败"),
		))
		return
	}

	dto.SuccessResponse(c, result)
}

// CreateArticle 创建文章
// @Summary 创建文章
// @Tags Article
// @Accept json
// @Produce json
// @Param request body dto.CreateArticleRequest true "创建文章请求"
// @Success 200 {object} response.Response
// @Router /articles [post]
func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	var req dto.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ValidationErrorResponse(c, err)
		return
	}

	userID, _ := c.Get("user_id")

	article, err := h.articleService.CreateArticle(req, userID.(uint))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("创建文章失败: "+err.Error()),
		))
		return
	}

	dto.SuccessResponse(c, article)
}

// GetArticle 获取文章详情
// @Summary 获取文章详情（包含当前版本内容）
// @Description 获取文章的基本信息、当前版本内容、标签、待审核提交等完整信息
// @Tags Article
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} response.Response{data=object{id=int,title=string,module_id=int,content=string,commit_message=string,version_number=int,current_version_id=int,current_user_role=string,is_review_required=bool,view_count=int,tags=[]string,pending_submissions=[]object,created_by=int,created_at=string,updated_at=string}}
// @Router /articles/{id} [get]
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	articleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的文章ID"),
		))
		return
	}

	// 获取用户ID（可选认证）
	var userID uint = 0
	if uid, exists := c.Get("user_id"); exists && uid != nil {
		userID = uid.(uint)
	}

	// 获取用户全局角色
	var userRole string
	if role, exists := c.Get("user_role"); exists && role != nil {
		userRole = role.(string)
	}

	article, err := h.articleService.GetArticle(uint(articleID), userID, userRole)
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取文章失败"),
		))
		return
	}

	dto.SuccessResponse(c, article)
}

// CreateSubmission 创建提交
// @Summary 提交文章修改
// @Tags Article
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Param request body dto.SubmissionRequest true "提交请求"
// @Success 200 {object} response.Response
// @Router /articles/{id}/submissions [post]
func (h *ArticleHandler) CreateSubmission(c *gin.Context) {
	articleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的文章ID"),
		))
		return
	}

	var req dto.SubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ValidationErrorResponse(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	var roleStr string
	if userRole != nil {
		roleStr = userRole.(string)
	}

	submission, publishedVersion, err := h.articleService.CreateSubmission(uint(articleID), req, userID.(uint), roleStr)
	if err != nil {
		// 检查是否是冲突错误
		if conflictErr, ok := err.(*MergeConflictError); ok {
			c.JSON(409, gin.H{
				"code":    40900,
				"message": "合并冲突",
				"data":    conflictErr.ConflictData,
			})
			return
		}
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("创建提交失败: "+err.Error()),
		))
		return
	}

	if submission == nil {
		if publishedVersion != nil {
			dto.SuccessResponse(c, gin.H{
				"message":           "修改已发布",
				"published":         true,
				"need_review":       false,
				"published_version": publishedVersion,
			})
			return
		}
		dto.SuccessResponse(c, gin.H{
			"message":     "修改已发布",
			"published":   true,
			"need_review": false,
		})
		return
	}

	// 需要审核，返回submission信息
	dto.SuccessResponse(c, gin.H{
		"message":     "提交成功，等待审核",
		"submission":  submission,
		"need_review": true,
	})
}

// GetReviews 获取审核列表
// @Summary 获取审核列表
// @Tags Review
// @Accept json
// @Produce json
// @Param status query string false "状态" Enums(pending, conflict_detected, all)
// @Param article_id query int false "文章ID"
// @Success 200 {object} response.Response
// @Router /reviews [get]
func (h *ArticleHandler) GetReviews(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")
	articleIDStr := c.Query("article_id")

	var articleID *uint
	if articleIDStr != "" {
		id, err := strconv.ParseUint(articleIDStr, 10, 32)
		if err == nil {
			idUint := uint(id)
			articleID = &idUint
		}
	}

	submissions, err := h.articleService.GetReviews(status, articleID)
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取审核列表失败"),
		))
		return
	}

	dto.SuccessResponse(c, submissions)
}

// GetReviewDetail 获取审核详情
// @Summary 获取审核详情（包含proposed_version完整信息）
// @Tags Review
// @Accept json
// @Produce json
// @Param id path int true "提交ID"
// @Success 200 {object} response.Response
// @Router /reviews/{id} [get]
func (h *ArticleHandler) GetReviewDetail(c *gin.Context) {
	submissionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的提交ID"),
		))
		return
	}

	// 获取用户ID（可选认证）
	var userID uint = 0
	if uid, exists := c.Get("user_id"); exists && uid != nil {
		userID = uid.(uint)
	}

	// 获取用户全局角色
	var userRole string
	if role, exists := c.Get("user_role"); exists && role != nil {
		userRole = role.(string)
	}

	detail, err := h.articleService.GetReviewDetail(uint(submissionID), userID, userRole)
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取审核详情失败"),
		))
		return
	}

	dto.SuccessResponse(c, detail)
}

// ReviewAction 审核操作
// @Summary 审核操作
// @Tags Review
// @Accept json
// @Produce json
// @Param id path int true "提交ID"
// @Param request body dto.ReviewActionRequest true "审核请求"
// @Success 200 {object} response.Response
// @Router /reviews/{id}/action [post]
func (h *ArticleHandler) ReviewAction(c *gin.Context) {
	submissionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的提交ID"),
		))
		return
	}

	var req dto.ReviewActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ValidationErrorResponse(c, err)
		return
	}

	reviewerID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	// 获取 userRole 字符串，如果不存在则为空字符串
	var roleStr string
	if userRole != nil {
		roleStr = userRole.(string)
	}

	result, err := h.articleService.ReviewSubmission(uint(submissionID), reviewerID.(uint), roleStr, req)
	if err != nil {
		// 检查是否是冲突错误
		if conflictErr, ok := err.(*MergeConflictError); ok {
			c.JSON(409, gin.H{
				"code":    409,
				"message": "检测到冲突",
				"data": gin.H{
					"conflict_data": conflictErr.ConflictData,
				},
			})
			return
		}
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("审核失败: "+err.Error()),
		))
		return
	}

	// 返回审核结果，包含完整的 published_version
	dto.SuccessResponse(c, result)
}

// UpdateBasicInfo 更新文章基础信息
// @Summary 更新文章基础信息
// @Description 更新文章的标题、标签、审核设置等基础信息（不涉及版本管理）
// @Tags Article
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Param request body dto.UpdateArticleBasicInfoRequest true "基础信息请求"
// @Success 200 {object} response.Response
// @Router /articles/{id}/basic-info [patch]
func (h *ArticleHandler) UpdateBasicInfo(c *gin.Context) {
	articleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的文章ID"),
		))
		return
	}

	var req dto.UpdateArticleBasicInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ValidationErrorResponse(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	var roleStr string
	if userRole != nil {
		roleStr = userRole.(string)
	}

	err = h.articleService.UpdateBasicInfo(uint(articleID), userID.(uint), roleStr, req)
	if err != nil {
		// 根据错误类型返回不同的响应
		if err.Error() == "permission denied: requires moderator or higher privileges" {
			dto.ErrorResponse(c, response.NewBusinessError(
				response.WithErrorCode(response.Forbidden),
				response.WithErrorMessage("权限不足：需要 moderator 或更高权限"),
			))
			return
		}

		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("更新基础信息失败: "+err.Error()),
		))
		return
	}

	dto.SuccessResponse(c, gin.H{"message": "更新成功"})
}

// AddCollaborator 添加协作者
// @Summary 添加协作者
// @Tags Article
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Param request body dto.AddCollaboratorRequest true "协作者请求"
// @Success 200 {object} response.Response
// @Router /articles/{id}/collaborators [post]
func (h *ArticleHandler) AddCollaborator(c *gin.Context) {
	articleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的文章ID"),
		))
		return
	}

	var req dto.AddCollaboratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ValidationErrorResponse(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	var roleStr string
	if userRole != nil {
		roleStr = userRole.(string)
	}

	err = h.articleService.AddCollaborator(uint(articleID), userID.(uint), roleStr, req)
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("添加协作者失败"),
		))
		return
	}

	dto.SuccessResponse(c, gin.H{"message": "添加成功"})
}

// GetVersions 获取文章版本列表
// @Summary 获取文章版本列表
// @Tags Article
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} response.Response
// @Router /articles/{id}/versions [get]
func (h *ArticleHandler) GetVersions(c *gin.Context) {
	articleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的文章ID"),
		))
		return
	}

	versions, err := h.articleService.GetVersions(uint(articleID))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取版本列表失败"),
		))
		return
	}

	dto.SuccessResponse(c, versions)
}

// GetVersion 获取特定版本
// @Summary 获取特定版本
// @Tags Article
// @Accept json
// @Produce json
// @Param id path int true "版本ID"
// @Success 200 {object} response.Response
// @Router /versions/{id} [get]
func (h *ArticleHandler) GetVersion(c *gin.Context) {
	versionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的版本ID"),
		))
		return
	}

	version, err := h.articleService.GetVersionByID(uint(versionID))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取版本失败"),
		))
		return
	}

	dto.SuccessResponse(c, version)
}

// GetVersionDiff 获取版本diff信息
// @Summary 获取版本diff信息
// @Description 获取指定版本和其基础版本的内容，用于前端展示diff
// @Tags 版本
// @Accept json
// @Produce json
// @Param id path int true "版本ID"
// @Success 200 {object} response.Response "成功返回版本diff信息"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "版本不存在"
// @Router /api/v1/versions/{id}/diff [get]
func (h *ArticleHandler) GetVersionDiff(c *gin.Context) {
	versionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的版本ID"),
		))
		return
	}

	diffData, err := h.articleService.GetVersionDiff(uint(versionID))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage(err.Error()),
		))
		return
	}

	dto.SuccessResponse(c, diffData)
}
