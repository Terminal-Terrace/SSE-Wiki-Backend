package main

import (
	"terminal-terrace/auth-service/config"
	"terminal-terrace/auth-service/internal/database"
	"terminal-terrace/auth-service/internal/route"

	_ "terminal-terrace/auth-service/docs" // Swagger 文档
)

// @title Auth Service API
// @version 1.0
// @description SSE-Wiki 认证服务 API 文档
// @termsOfService https://github.com/Terminal-Terrace/SSE-Wiki-Backend

// @contact.name API Support
// @contact.url https://github.com/Terminal-Terrace/SSE-Wiki-Backend/issues
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	config.MustLoad("config.yaml")
	database.InitDatabase()
	r := route.SetupRouter()

	r.Run(":8081")
}