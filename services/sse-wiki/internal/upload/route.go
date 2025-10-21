package upload

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	h := NewHandler(db)

	g := r.Group("/upload")
	{
		g.POST("/init", h.Init)
		g.POST("/chunk", h.Chunk)
		g.POST("/complete", h.Complete)
	}
}


