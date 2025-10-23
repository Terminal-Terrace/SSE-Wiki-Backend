package preference

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupPreferRouter(r *gin.RouterGroup, db *gorm.DB) {
	handler := NewPreferHandler(db, 0.95)

	r.GET("/get_prefer", handler.GetPrefer)
	r.GET("/get_best_prefer", handler.GetBestPrefer)

	r.POST("/update_prefer", handler.UpdatePreference)
	r.POST("/set_prefer", handler.SetPreference)
}
