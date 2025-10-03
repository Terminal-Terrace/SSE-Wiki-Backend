package prelogin

import (
	"terminal-terrace/auth-service/internal/dto"
	"terminal-terrace/auth-service/internal/pkg"
	"terminal-terrace/response"

	"github.com/gin-gonic/gin"
)

type PreLoginHandler struct{}

func (h *PreLoginHandler) handle(c *gin.Context) {
	// 解析参数
	var req PreLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("请检查参数"),
		))
		return
	}

	// 生成 state
	state, err := pkg.GenerateState()
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("生成 state 失败"),
		))
		return
	}

	// 保存 state 和重定向地址到 Redis
	if err := pkg.SaveStateWithRedirect(state, req.RedirectUrl); err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("保存 state 失败"),
		))
		return
	}

	// 返回 state
	dto.SuccessResponse(c, PreLoginResponse{
		State: state,
	})
}