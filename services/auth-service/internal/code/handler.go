package code

import (
	"terminal-terrace/auth-service/internal/dto"
	"terminal-terrace/response"

	"github.com/gin-gonic/gin"
)

type CodeHandler struct {
	service *CodeService
}

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
