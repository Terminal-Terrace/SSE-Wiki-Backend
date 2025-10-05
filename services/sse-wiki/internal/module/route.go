package module

import (
	"github.com/gin-gonic/gin"
	"terminal-terrace/sse-wiki/internal/middleware"
)

func RegisterRoutes(r *gin.RouterGroup) {
	moduleHandler := NewModuleHandler()
	lockHandler := NewLockHandler()

	modules := r.Group("/modules")
	{
		// 查询类接口（可选认证：任何人都可以访问，但如果有token会解析用户信息）
		modules.GET("", middleware.OptionalJWTAuth(), moduleHandler.GetModuleTree)
		modules.GET("/:id", middleware.OptionalJWTAuth(), moduleHandler.GetModule)
		modules.GET("/:id/breadcrumbs", middleware.OptionalJWTAuth(), moduleHandler.GetBreadcrumbs)
		modules.GET("/:id/moderators", middleware.OptionalJWTAuth(), moduleHandler.GetModerators)

		// 编辑类接口（必需认证）
		authRequired := modules.Group("")
		authRequired.Use(middleware.JWTAuth())
		{
			authRequired.POST("", moduleHandler.CreateModule)
			authRequired.PUT("/:id", moduleHandler.UpdateModule)
			authRequired.DELETE("/:id", moduleHandler.DeleteModule)

			// 编辑锁
			authRequired.POST("/lock", lockHandler.HandleLock)

			// 协作者管理
			authRequired.POST("/:id/moderators", moduleHandler.AddModerator)
			authRequired.DELETE("/:id/moderators/:userId", moduleHandler.RemoveModerator)
		}
	}
}
