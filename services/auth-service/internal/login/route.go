package login

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	h := &LoginHandler{}
	r.POST("/login", h.handle)
}