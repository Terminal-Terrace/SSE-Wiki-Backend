package me

import (
	"terminal-terrace/auth-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	handler := &MeHandler{}

	// 需要认证的接口
	r.GET("/me", middleware.JWTAuth(), handler.GetCurrentUser)
}
