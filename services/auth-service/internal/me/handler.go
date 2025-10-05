package me

import (
	"terminal-terrace/auth-service/internal/dto"
	"terminal-terrace/response"

	"github.com/gin-gonic/gin"
)

type MeHandler struct{}

// GetCurrentUser 获取当前登录用户信息
// @Summary 获取当前用户信息
// @Description 从 Cookie 中的 access_token 获取当前登录用户信息
// @Tags 认证
// @Produce json
// @Success 200 {object} UserInfoResponse
// @Router /auth/me [get]
func (h *MeHandler) GetCurrentUser(c *gin.Context) {
	// 从上下文获取用户信息（由中间件设置）
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Unauthorized),
			response.WithErrorMessage("未登录"),
		))
		return
	}

	username, _ := c.Get("username")
	email, _ := c.Get("email")
	role, _ := c.Get("user_role")

	dto.SuccessResponse(c, gin.H{
		"user_id":  userID,
		"username": username,
		"email":    email,
		"role":     role,
	})
}
