package handler

import (
	"terminal-terrace/sse-wiki/internal/service"

	"github.com/gin-gonic/gin"
	"terminal-terrace/sse-wiki/internal/dto"
)

type ExampleHandler struct {
	// TODO: 添加需要的依赖
	exampleService *service.ExampleService
}

func NewExampleHandler(
	// TODO: 增加需要的依赖
	exampleService *service.ExampleService,
) *ExampleHandler {
	return &ExampleHandler{
		// TODO: 增加需要的依赖
		exampleService: exampleService,
	}
}

func (h *ExampleHandler) HandleGood(c *gin.Context) {
	result, err := h.exampleService.DoSomeGood()
	if err != nil {
		// TODO: 处理错误
		dto.ErrorResponse(c, err)
		return
	}
	// TODO: 处理成功结果
	dto.SuccessResponse(c, result)
}

func (h *ExampleHandler) HandleBad(c *gin.Context) {
	result, err := h.exampleService.DoSomeBad()
	if err != nil {
		// TODO: 处理错误
		dto.ErrorResponse(c, err)
		return
	}
	// TODO: 处理成功结果
	dto.SuccessResponse(c, result)
}
