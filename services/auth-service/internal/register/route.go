package register

import (
	"github.com/gin-gonic/gin"
	"terminal-terrace/auth-service/internal/database"
	"terminal-terrace/auth-service/internal/refresh"
)

func RegisterRoutes(r *gin.RouterGroup) {
	repo := refresh.NewRefreshTokenRepository(database.RedisDB)
	service := NewRegisterService(repo)
	h := &RegisterHandler{
		service: service,
	}
	r.POST("/register", h.handle)
}