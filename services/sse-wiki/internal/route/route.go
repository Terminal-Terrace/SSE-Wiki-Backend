package route

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"terminal-terrace/sse-wiki/internal/handler"
	"terminal-terrace/sse-wiki/internal/module"
	"terminal-terrace/sse-wiki/internal/service"
)

func initRoute(r *gin.Engine) {
	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 初始化依赖
	exampleService := service.NewExampleService()

	// 初始化handler
	exampleHandler := handler.NewExampleHandler(exampleService)

	// API 路由组
	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/good", exampleHandler.HandleGood)
		apiV1.GET("/bad", exampleHandler.HandleBad)
	}

	// 模块管理路由
	module.RegisterRoutes(apiV1)
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 支持多个前端域名（开发环境）
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:3001",
	}

	// 如果设置了环境变量，则使用环境变量指定的域名
	if origin := os.Getenv("FRONTEND_URL"); origin != "" {
		allowedOrigins = []string{origin}
	}

	// 设置跨域请求
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true, // 允许携带 cookie
	}))

	initRoute(r)

	return r
}
