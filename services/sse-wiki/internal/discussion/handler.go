package discussion

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DiscussionHandler 讨论区处理器
type DiscussionHandler struct {
	service DiscussionService
}

// NewDiscussionHandler 创建处理器实例
func NewDiscussionHandler(service DiscussionService) *DiscussionHandler {
	return &DiscussionHandler{
		service: service,
	}
}

// GetArticleComments 获取文章的所有评论
// GET /api/articles/:id/discussions
func (h *DiscussionHandler) GetArticleComments(c *gin.Context) {
	// 1. 获取文章ID
	// 同步修改：路由已改为 :id，因此这里需要从 c.Param("id") 获取
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid article ID",
			Error:   err.Error(),
		})
		return
	}

	// 2. 调用服务层
	result, err := h.service.GetArticleComments(uint(articleID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 3. 返回成功响应
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    result,
	})
}

// CreateComment 创建顶级评论
// POST /api/articles/:id/discussions
func (h *DiscussionHandler) CreateComment(c *gin.Context) {
	// 1. 获取文章ID
	// 同步修改：路由已改为 :id，因此这里需要从 c.Param("id") 获取
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid article ID",
			Error:   err.Error(),
		})
		return
	}

	// 2. 获取当前用户ID（从认证中间件获取）
	userID, exists := c.Get("user_id")
	if !exists {
		// 临时方案：如果没有认证系统，使用默认用户ID 1
		// TODO: 集成认证系统后，应该返回 401 错误
		userID = uint(1)
		// c.JSON(http.StatusUnauthorized, ErrorResponse{
		//  Code:    http.StatusUnauthorized,
		//  Message: "Unauthorized",
		//  Error:   "User not authenticated",
		// })
		// return
	}

	// 3. 绑定请求参数
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// 4. 调用服务层
	result, err := h.service.CreateComment(uint(articleID), userID.(uint), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 5. 返回成功响应
	c.JSON(http.StatusCreated, Response{
		Code:    http.StatusCreated,
		Message: "Comment created successfully",
		Data:    result,
	})
}

// ReplyComment 回复评论
// POST /api/discussions/:commentId/replies
func (h *DiscussionHandler) ReplyComment(c *gin.Context) {
	// 1. 获取父评论ID
	commentID, err := strconv.ParseUint(c.Param("commentId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid comment ID",
			Error:   err.Error(),
		})
		return
	}

	// 2. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		userID = uint(1)
		// c.JSON(http.StatusUnauthorized, ErrorResponse{
		// 	Code:    http.StatusUnauthorized,
		// 	Message: "Unauthorized",
		// 	Error:   "User not authenticated",
		// })
		// return
	}

	// 3. 绑定请求参数
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// 4. 调用服务层
	result, err := h.service.ReplyComment(uint(commentID), userID.(uint), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 5. 返回成功响应
	c.JSON(http.StatusCreated, Response{
		Code:    http.StatusCreated,
		Message: "Reply created successfully",
		Data:    result,
	})
}

// UpdateComment 更新评论
// PUT /api/discussions/:commentId
func (h *DiscussionHandler) UpdateComment(c *gin.Context) {
	// 1. 获取评论ID
	commentID, err := strconv.ParseUint(c.Param("commentId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid comment ID",
			Error:   err.Error(),
		})
		return
	}

	// 2. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		userID = uint(1)
		// c.JSON(http.StatusUnauthorized, ErrorResponse{
		// 	Code:    http.StatusUnauthorized,
		// 	Message: "Unauthorized",
		// 	Error:   "User not authenticated",
		// })
		// return
	}

	// 3. 绑定请求参数
	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// 4. 调用服务层
	result, err := h.service.UpdateComment(uint(commentID), userID.(uint), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 5. 返回成功响应
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "Comment updated successfully",
		Data:    result,
	})
}

// DeleteComment 删除评论
// DELETE /api/discussions/:commentId
func (h *DiscussionHandler) DeleteComment(c *gin.Context) {
	// 1. 获取评论ID
	commentID, err := strconv.ParseUint(c.Param("commentId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid comment ID",
			Error:   err.Error(),
		})
		return
	}

	// 2. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		userID = uint(1)
		// c.JSON(http.StatusUnauthorized, ErrorResponse{
		// 	Code:    http.StatusUnauthorized,
		// 	Message: "Unauthorized",
		// 	Error:   "User not authenticated",
		// })
		// return
	}

	// 3. 调用服务层
	err = h.service.DeleteComment(uint(commentID), userID.(uint))
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 4. 返回成功响应
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "Comment deleted successfully",
	})
}

// ========== 错误处理 ==========

// handleError 统一错误处理
func (h *DiscussionHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Resource not found",
			Error:   err.Error(),
		})
	case errors.Is(err, ErrCommentNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Comment not found",
			Error:   err.Error(),
		})
	case errors.Is(err, ErrDiscussionNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Discussion not found",
			Error:   err.Error(),
		})
	case errors.Is(err, ErrUnauthorized):
		c.JSON(http.StatusForbidden, ErrorResponse{
			Code:    http.StatusForbidden,
			Message: "Permission denied",
			Error:   err.Error(),
		})
	case errors.Is(err, ErrInvalidParentID):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid parent comment",
			Error:   err.Error(),
		})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Error:   err.Error(),
		})
	}
}
