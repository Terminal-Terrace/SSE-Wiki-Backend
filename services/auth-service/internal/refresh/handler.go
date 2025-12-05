package refresh

import (
	"terminal-terrace/auth-service/internal/dto"
	"terminal-terrace/response"

	"github.com/gin-gonic/gin"
)

type RefreshTokenHandler struct {
	service *RefreshTokenService
}

func NewRefreshTokenHandler(service *RefreshTokenService) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		service: service,
	}
}

// Handle 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 使用 Cookie 中的刷新令牌获取新的访问令牌，新的刷新令牌会自动更新到 Cookie 中
// @Tags 认证
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=RefreshTokenResponse} "成功返回新的访问令牌"
// @Failure 200 {object} dto.Response "刷新令牌无效或已过期"
// @Router /auth/refresh [post]
func (h *RefreshTokenHandler) Handle(c *gin.Context) {
	// 从 cookie 中获取 refresh token
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("未找到刷新令牌"),
		))
		return
	}

	// 调用服务层
	result, bizErr := h.service.RefreshToken(RefreshTokenRequest{RefreshToken: refreshToken})
	if bizErr != nil {
		dto.ErrorResponse(c, bizErr)
		return
	}

	// 设置新的 refresh token 到 cookie (httpOnly)
	c.SetCookie("refresh_token", result.NewRefreshToken, 3600*24*7, "/", "", false, true)

	// 只返回 access token（refresh token 不暴露给前端）
	dto.SuccessResponse(c, RefreshTokenResponse{
		AccessToken: result.AccessToken,
	})
}
