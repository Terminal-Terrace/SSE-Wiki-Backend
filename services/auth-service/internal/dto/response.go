package dto

import (
	res "terminal-terrace/response"

	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, data any) {
	c.JSON(200, res.SuccessResponse(data))
}

func ErrorResponse(c *gin.Context, err *res.BusinessError) {
	c.JSON(200, res.ErrorResponse(err.Code, err.Msg))
}