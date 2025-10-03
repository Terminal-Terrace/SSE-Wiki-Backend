package register

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	h := &RegisterHandler{}
	r.POST("/register", h.handle)
}