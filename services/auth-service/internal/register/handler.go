package register

import (
	"terminal-terrace/auth-service/internal/dto"
	"terminal-terrace/response"

	"github.com/gin-gonic/gin"
)

type RegisterHandler struct {
	service *RegisterService
}

// Handle 用户注册
// @Summary 用户注册
// @Description 注册新用户账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册请求"
// @Success 200 {object} dto.Response{data=map[string]string} "注册成功，返回重定向 URL"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Failure 409 {object} dto.Response "用户名或邮箱已存在"
// @Failure 500 {object} dto.Response "服务器内部错误"
// @Router /auth/register [post]
func (h *RegisterHandler) handle(c *gin.Context) {
	// 解析参数
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("请检查参数"),
		))
		return
	}

	// 调用注册服务
	result, err := h.service.Register(req)
	if err != nil {
		dto.ErrorResponse(c, err)
		return
	}

	// 设置 Cookie
	c.SetCookie("refresh_token", result.RefreshToken, 3600*24*7, "/", "", false, true)

	dto.SuccessResponse(c, gin.H{
		"redirect_url": result.RedirectUrl,
	})
}