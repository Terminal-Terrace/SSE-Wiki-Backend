package preference

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupPreferRouter(r *gin.RouterGroup, db *gorm.DB) {
	handler := NewPreferHandler(db, 0.95)
	r.POST("/get_refer", handler.GetPrefer)
	r.POST("/get_best_refer", handler.GetBestRefer)
	r.POST("/update_refer", handler.UpdatePreference)
}
