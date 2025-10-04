package route

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func initRoute(r *gin.Engine) {
	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API 路由组
	apiV1 := r.Group("/api/v1")
	{
		// TODO: 在这里添加你的路由
		// 示例：
		// apiV1.GET("/example", handler.Example)
		_ = apiV1
	}
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	origin := os.Getenv("FRONTEND_URL")
	if origin == "" {
		origin = "http://localhost:5173" // 默认值
	}

	// 设置跨域请求
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{origin},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	initRoute(r)

	return r
}
