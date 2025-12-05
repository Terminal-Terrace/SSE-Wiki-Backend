package discussion

import (
	"log"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupDiscussionRoutes 注册讨论区相关路由（不带用户服务）
// 评论不会显示用户信息，仅显示 user_id
func SetupDiscussionRoutes(router *gin.RouterGroup, db *gorm.DB) {
	userService := NewSimpleUserService(db)
	SetupDiscussionRoutesWithUserService(router, db, userService)
}

// SetupDiscussionRoutesWithUserService 注册讨论区相关路由（带用户服务）
// 如果提供了 userService，评论会包含用户的详细信息
func SetupDiscussionRoutesWithUserService(router *gin.RouterGroup, db *gorm.DB, userService UserService) {
	// 1. 创建 Repository 层
	repo := NewDiscussionRepository(db)
	if repo == nil {
		log.Fatal("Failed to create discussion repository")
	}

	service := NewDiscussionService(repo, db, userService)
	if service == nil {
		log.Fatal("Failed to create discussion service")
	}

	handler := NewDiscussionHandler(service)
	if handler == nil {
		log.Fatal("Failed to create discussion handler")
	}

	// 4. 注册路由
	// ========== 文章评论路由 ==========
	articles := router.Group("/articles")
	{
		articles.GET("/:id/discussions", handler.GetArticleComments)

		// 发表新评论（TODO: 需要登录，暂时不添加中间件）
		articles.POST("/:id/discussions", handler.CreateComment)
	}

	// ========== 评论操作路由 ==========
	discussions := router.Group("/discussions")
	{
		// 回复评论（TODO: 需要登录）
		discussions.POST("/:commentId/replies", handler.ReplyComment)

		// 编辑评论（TODO: 需要登录且只能编辑自己的评论）
		discussions.PUT("/:commentId", handler.UpdateComment)

		// 删除评论（TODO: 需要登录且只能删除自己的评论）
		discussions.DELETE("/:commentId", handler.DeleteComment)
	}
}

// RegisterRoutesSimple 简化版路由注册（如果不需要认证中间件）
func RegisterRoutesSimple(router *gin.RouterGroup, handler *DiscussionHandler) {
	api := router.Group("/api/v1")

	// 文章评论
	articles := api.Group("/articles")
	{
		// 统一使用 :id
		articles.GET("/:id/discussions", handler.GetArticleComments)
		articles.POST("/:id/discussions", handler.CreateComment)
	}

	// 评论操作
	discussions := api.Group("/discussions")
	{
		discussions.POST("/:commentId/replies", handler.ReplyComment)
		discussions.PUT("/:commentId", handler.UpdateComment)
		discussions.DELETE("/:commentId", handler.DeleteComment)
	}
}
