package register

import (
	"terminal-terrace/auth-service/internal/dto"
	"terminal-terrace/response"

	"github.com/gin-gonic/gin"
)

type RegisterHandler struct {
	service *RegisterService
}

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