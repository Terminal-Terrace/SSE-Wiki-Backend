package dto

import (
	res "terminal-terrace/response"

	"github.com/gin-gonic/gin"
)

// Response 统一响应格式
type Response struct {
	Code    int    `json:"code" example:"100"`              // 状态码：100-成功，其他-失败
	Message string `json:"message" example:"success"`       // 响应消息
	Data    any    `json:"data,omitempty"`                  // 响应数据
}

func SuccessResponse(c *gin.Context, data any) {
	c.JSON(200, res.SuccessResponse(data))
}

func ErrorResponse(c *gin.Context, err *res.BusinessError) {
	c.JSON(200, res.ErrorResponse(err.Code, err.Msg))
}