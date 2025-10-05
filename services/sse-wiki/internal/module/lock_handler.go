package module

import (
	"terminal-terrace/response"
	"terminal-terrace/sse-wiki/internal/database"
	"terminal-terrace/sse-wiki/internal/dto"

	"github.com/gin-gonic/gin"
)

type LockHandler struct {
	lockService *LockService
}

func NewLockHandler() *LockHandler {
	return &LockHandler{
		lockService: NewLockService(database.RedisDB),
	}
}

// HandleLock 处理编辑锁（获取/释放）
// @Summary 编辑锁操作
// @Description 获取或释放全局编辑锁
// @Tags 模块管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body LockRequest true "锁操作请求"
// @Success 200 {object} response.Response{data=LockResponse}
// @Router /modules/lock [post]
func (h *LockHandler) HandleLock(c *gin.Context) {
	var req LockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("参数错误"),
		))
		return
	}

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	if req.Action == "acquire" {
		// 获取锁
		lockResp, err := h.lockService.AcquireLock(userID.(uint), username.(string))
		if err != nil {
			dto.ErrorResponse(c, err.(*response.BusinessError))
			return
		}
		dto.SuccessResponse(c, lockResp)
	} else if req.Action == "release" {
		// 释放锁
		if err := h.lockService.ReleaseLock(userID.(uint)); err != nil {
			dto.ErrorResponse(c, err.(*response.BusinessError))
			return
		}
		dto.SuccessResponse(c, gin.H{"message": "释放成功"})
	}
}
