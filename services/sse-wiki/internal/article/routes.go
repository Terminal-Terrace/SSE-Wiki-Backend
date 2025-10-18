package article

import (
	"terminal-terrace/sse-wiki/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupArticleRoutes 设置文章相关路由
func SetupArticleRoutes(r *gin.RouterGroup, db *gorm.DB) {
	// 初始化handler（内部会自动初始化所有依赖）
	articleHandler := NewArticleHandler(db)

	// 模块路由
	modules := r.Group("/modules")
	{
		modules.GET("/:id/articles", articleHandler.GetArticlesByModule) // 获取模块下的文章列表
	}

	// 文章路由 - 需要认证
	articlesAuth := r.Group("/articles")
	articlesAuth.Use(middleware.JWTAuth()) // 需要认证
	{
		articlesAuth.POST("", articleHandler.CreateArticle)                     // 创建文章（需要认证）
		articlesAuth.POST("/:id/submissions", articleHandler.CreateSubmission)  // 提交修改（需要认证）
		articlesAuth.PATCH("/:id/basic-info", articleHandler.UpdateBasicInfo)   // 更新基础信息（需要认证）
		articlesAuth.POST("/:id/collaborators", articleHandler.AddCollaborator) // 添加协作者（需要认证）
	}

	// 文章路由 - 可选认证（用于获取用户角色信息）
	articlesOptional := r.Group("/articles")
	articlesOptional.Use(middleware.OptionalJWTAuth()) // 可选认证
	{
		articlesOptional.GET("/:id", articleHandler.GetArticle)           // 获取文章详情（可选认证）
		articlesOptional.GET("/:id/versions", articleHandler.GetVersions) // 获取版本列表（可选认证）
	}

	// 版本路由
	versions := r.Group("/versions")
	{
		versions.GET("/:id", articleHandler.GetVersion)          // 获取特定版本内容
		versions.GET("/:id/diff", articleHandler.GetVersionDiff) // 获取版本diff信息
	}

	// 审核路由 - 需要认证
	reviews := r.Group("/reviews")
	reviews.Use(middleware.JWTAuth()) // 需要认证
	{
		reviews.GET("", articleHandler.GetReviews)               // 获取审核列表（需要认证）
		reviews.GET("/:id", articleHandler.GetReviewDetail)      // 获取审核详情（需要认证）
		reviews.POST("/:id/action", articleHandler.ReviewAction) // 审核操作（需要认证）
	}
}
