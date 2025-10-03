package login

import (
	"terminal-terrace/auth-service/internal/dto"
	"terminal-terrace/response"

	"github.com/gin-gonic/gin"
)

func Handler(c *gin.Context) {
	// 解析参数
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(response.WithErrorCode(response.ParseError), response.WithErrorMessage("请检查参数")))
		return
	}

	// TODO: CheckState

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

	// setCookie
	// TODO: 配置cookie
	c.SetCookie("refresh_token", result.RefreshToken, 3600*24*7, "/", "", false, true)
	dto.SuccessResponse(c, gin.H{
		"redirect_url": result.RedirectUrl,
	})
}