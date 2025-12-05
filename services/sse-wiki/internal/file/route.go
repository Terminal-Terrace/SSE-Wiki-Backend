package file

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) {
	handler := NewHandler(db)

	fileGroup := router.Group("/files")
	{
		// 在线预览文件
		fileGroup.GET("/:id", handler.GetFile)
		
		// 强制下载文件
		fileGroup.GET("/:id/download", handler.DownloadFile)
		
		// 获取文件元数据（不返回文件内容）
		fileGroup.GET("/:id/info", handler.GetFileInfo)
	}
}



