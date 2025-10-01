package route

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"os"
	"terminal-terrace/sse-wiki/internal/handler"
	"terminal-terrace/sse-wiki/internal/service"
)

func initRoute(r *gin.Engine) {
	// 初始化依赖
	exampleService := service.NewExampleService()

	// 初始化handler
	exampleHandler := handler.NewExampleHandler(exampleService)

	r.GET("/good", exampleHandler.HandleGood)
	r.GET("/bad", exampleHandler.HandleBad)
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
