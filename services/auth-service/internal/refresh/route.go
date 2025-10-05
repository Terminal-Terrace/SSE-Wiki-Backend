package refresh

import (
	"github.com/gin-gonic/gin"
	"terminal-terrace/auth-service/internal/database"
)

func RegisterRoutes(r *gin.RouterGroup) {
	repo := NewRefreshTokenRepository(database.RedisDB)
	service := NewRefreshTokenService(repo)
	handler := NewRefreshTokenHandler(service)
	r.POST("/refresh", handler.Handle)
}
