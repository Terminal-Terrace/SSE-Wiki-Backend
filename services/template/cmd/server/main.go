package main

import (
	"terminal-terrace/template/config"
	"terminal-terrace/template/internal/database"
	"terminal-terrace/template/internal/route"

	_ "terminal-terrace/template/docs" // Swagger 文档
)

// @title Template Service API
// @version 1.0
// @description Template 服务 API 文档模板
// @termsOfService https://github.com/Terminal-Terrace/SSE-Wiki-Backend

// @contact.name API Support
// @contact.url https://github.com/Terminal-Terrace/SSE-Wiki-Backend/issues
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8082
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
	r.Run(":8082")
}
