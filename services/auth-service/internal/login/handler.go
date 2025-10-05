package login

import (
	"terminal-terrace/auth-service/internal/dto"
	"terminal-terrace/response"

	"github.com/gin-gonic/gin"
)

type LoginHandler struct{}

// Handle 用户登录
// @Summary 用户登录
// @Description 支持多种登录方式：SSE-Wiki 账号密码登录、GitHub OAuth、SSE-Market OAuth
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Router /auth/login [post]
func (h *LoginHandler) handle(c *gin.Context) {
	// 解析参数
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(response.WithErrorCode(response.ParseError), response.WithErrorMessage("请检查参数")))
		return
	}

	// 根据登录类型选择对应的服务
	service, exists := loginServices[req.Type]
	if !exists {
		// TODO: 写一个更合适的错误码
		dto.ErrorResponse(c, response.NewBusinessError(response.WithErrorCode(response.Fail), response.WithErrorMessage("不支持的登录类型")))
		return
	}

	// 调用对应的登录服务
	result, err := service.Login(req)
	if err != nil {
		dto.ErrorResponse(c, err)
		return
	}

	// 设置 access_token 和 refresh_token 到 Cookie（HttpOnly + Secure）
	// access_token: 15分钟有效期
	c.SetCookie("access_token", result.AccessToken, 15*60, "/", "", false, true)
	// refresh_token: 7天有效期
	c.SetCookie("refresh_token", result.RefreshToken, 3600*24*7, "/", "", false, true)

	// 响应体只返回 redirect_url
	dto.SuccessResponse(c, gin.H{
		"redirect_url": result.RedirectUrl,
	})
}
