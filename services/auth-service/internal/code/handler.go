package code

import (
	"terminal-terrace/auth-service/internal/dto"
	"terminal-terrace/response"

	"github.com/gin-gonic/gin"
)

type CodeHandler struct {
	service *CodeService
}

// handle 发送验证码
// @Summary 发送验证码
// @Description 发送邮箱验证码用于注册或重置密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body SendCodeRequest true "发送验证码请求"
// @Success 200 {object} dto.Response "发送成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Failure 500 {object} dto.Response "服务器内部错误"
// @Router /auth/code [post]
func (h *CodeHandler) handle(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(response.WithErrorCode(response.ParseError), response.WithErrorMessage(err.Error())))
		return
	}

	err := h.service.SendCode(req)

	if err != nil {
		dto.ErrorResponse(c, err)
		return
	}

	dto.SuccessResponse(c, nil)
}
