package logout

import (
	"terminal-terrace/auth-service/internal/dto"

	"github.com/gin-gonic/gin"
)

type LogoutHandler struct{}

// Logout 用户退出登录
// @Summary 用户退出登录
// @Description 清除 access_token 和 refresh_token Cookie，退出登录
// @Tags 认证
// @Produce json
// @Success 200 {object} response.Response
// @Router /auth/logout [post]
func (h *LogoutHandler) Logout(c *gin.Context) {
	// 清除 access_token Cookie
	c.SetCookie(
		"access_token",
		"",
		-1,  // 立即过期
		"/",
		"",
		false,
		true,
	)

	// 清除 refresh_token Cookie
	c.SetCookie(
		"refresh_token",
		"",
		-1,  // 立即过期
		"/",
		"",
		false,
		true,
	)

	// TODO: 可选 - 从 Redis 删除 refresh_token
	// 如果需要更彻底的退出（让所有设备登出），可以在这里删除 Redis 中的 refresh_token

	dto.SuccessResponse(c, gin.H{
		"message": "退出成功",
	})
}
