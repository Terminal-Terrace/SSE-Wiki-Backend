package main

import (
	"terminal-terrace/sse-wiki/config"
	"terminal-terrace/sse-wiki/internal/database"
	"terminal-terrace/sse-wiki/internal/route"

	_ "terminal-terrace/sse-wiki/docs"
)

// @title SSE-Wiki Service API
// @version 1.0
// @description SSE-Wiki 服务 API 文档
// @termsOfService https://github.com/Terminal-Terrace/SSE-Wiki-Backend

// @contact.name API Support
// @contact.url https://github.com/Terminal-Terrace/SSE-Wiki-Backend/issues
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. 加载配置
	config.MustLoad("config.yaml")

	// 2. 初始化数据库
	database.InitDatabase()

	// 3. 设置路由
	r := route.SetupRouter()

	// 4. 启动服务
	r.Run(":8080")
}