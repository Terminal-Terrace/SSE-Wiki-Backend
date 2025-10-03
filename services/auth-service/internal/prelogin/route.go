package prelogin

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	h := &PreLoginHandler{}
	r.POST("/pre-login", h.handle)
}