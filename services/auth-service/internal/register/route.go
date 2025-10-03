package register

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	service := &RegisterService{}
	h := &RegisterHandler{
		service: service,
	}
	r.POST("/register", h.handle)
}