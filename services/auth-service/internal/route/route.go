package route

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"terminal-terrace/auth-service/internal/code"
	"terminal-terrace/auth-service/internal/login"
	"terminal-terrace/auth-service/internal/prelogin"
	"terminal-terrace/auth-service/internal/refresh"
	"terminal-terrace/auth-service/internal/register"
)

func initRoute(r *gin.Engine) {
	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API 路由组
	apiV1 := r.Group("/api/v1")
	{
		authGroup := apiV1.Group("/auth")
		prelogin.RegisterRoutes(authGroup)
		login.RegisterRoutes(authGroup)
		register.RegisterRoutes(authGroup)
		code.RegisterRoutes(authGroup)
		refresh.RegisterRoutes(authGroup)
	}
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 允许多个前端端口
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:3001",
	}

	// 如果设置了环境变量，添加到允许列表
	if envOrigin := os.Getenv("FRONTEND_URL"); envOrigin != "" {
		allowedOrigins = append(allowedOrigins, envOrigin)
	}

	// 设置跨域请求
	r.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	initRoute(r)

	return r
}
