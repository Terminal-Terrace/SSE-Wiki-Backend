package main

import (
	"fmt"
	"log"

	"terminal-terrace/auth-service/config"
	"terminal-terrace/auth-service/internal/database"
	"terminal-terrace/auth-service/internal/model"
	"terminal-terrace/auth-service/internal/route"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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
	// 1. 加载配置
	config.MustLoad("config.yaml")

	// 2. 确保数据库存在
	if err := ensureDatabaseExists(); err != nil {
		log.Fatalf("数据库创建失败: %v", err)
	}

	// 3. 初始化数据库连接
	database.InitDatabase()

	// 4. 初始化数据库表
	if err := model.InitTable(database.GetDB()); err != nil {
		log.Fatalf("数据库表初始化失败: %v", err)
	}

	// 5. 设置路由
	r := route.SetupRouter()

	// 6. 启动服务
	log.Printf("[auth-service] 服务启动在端口 :8081")
	r.Run(":8081")
}

// ensureDatabaseExists 确保数据库存在，如果不存在则创建
func ensureDatabaseExists() error {
	databaseConf := config.Conf.Database

	// 首先连接到postgres数据库（默认数据库）
	dsn := fmt.Sprintf("host=%s user=%s password=%s port=%d sslmode=disable",
		databaseConf.Host, databaseConf.Username, databaseConf.Password, databaseConf.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接到PostgreSQL失败: %v", err)
	}

	// 检查数据库是否存在
	var exists bool
	checkSQL := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)"
	if err = db.Raw(checkSQL, databaseConf.Database).Scan(&exists).Error; err != nil {
		return fmt.Errorf("检查数据库是否存在失败: %v", err)
	}

	if !exists {
		log.Printf("[auth-service] 数据库 '%s' 不存在，正在创建...", databaseConf.Database)
		createSQL := fmt.Sprintf("CREATE DATABASE %s", databaseConf.Database)
		if err = db.Exec(createSQL).Error; err != nil {
			return fmt.Errorf("创建数据库失败: %v", err)
		}
		log.Printf("[auth-service] 数据库 '%s' 创建成功", databaseConf.Database)
	}

	// 关闭连接
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
