package main

import (
	"fmt"
	"log"

	"terminal-terrace/auth-service/config"
	"terminal-terrace/auth-service/internal/database"
	grpcserver "terminal-terrace/auth-service/internal/grpc"
	"terminal-terrace/auth-service/internal/model"
	"terminal-terrace/auth-service/internal/route"
	"terminal-terrace/email"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	// _ "terminal-terrace/auth-service/docs" // Swagger 文档 (暂时禁用)
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

	// 3.1 同步数据库结构
	if err := model.InitTable(database.GetDB()); err != nil {
		log.Fatalf("数据库表初始化失败: %v", err)
	}

	// // 3.2 更新 Swagger 文档
	// if err := refreshSwaggerDocs(); err != nil {
	// 	log.Printf("[auth-service] Swagger 文档更新失败: %v", err)
	// }

	// 4. 初始化邮件客户端（gRPC 服务需要）
	mailer := email.NewClient(&config.Conf.Smtp)

	// 5. 启动 gRPC server (goroutine)
	go func() {
		grpcPort := config.Conf.GRPC.Port
		if grpcPort == 0 {
			grpcPort = 50051 // 默认端口
		}
		authService := grpcserver.NewAuthServiceImpl(mailer)
		server, err := grpcserver.NewServer(grpcPort, authService)
		if err != nil {
			log.Fatalf("[auth-service] gRPC server 启动失败: %v", err)
		}
		log.Printf("[auth-service] gRPC server 启动在端口 :%d", grpcPort)
		if err := server.Start(); err != nil {
			log.Fatalf("[auth-service] gRPC server 运行失败: %v", err)
		}
	}()

	// 6. 设置 REST 路由
	r := route.SetupRouter()

	// 7. 启动 REST server (blocking)
	log.Printf("[auth-service] REST server 启动在端口 :8081")
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

// // refreshSwaggerDocs 生成最新的 Swagger 文档
// func refreshSwaggerDocs() error {
// 	serviceRoot, err := resolveServiceRoot()
// 	if err != nil {
// 		return err
// 	}

// 	args := []string{
// 		"init",
// 		"-g", filepath.ToSlash(filepath.Join("cmd", "server", "main.go")),
// 		"-o", filepath.ToSlash("docs"),
// 		"--parseDependency",
// 		"--parseInternal",
// 	}

// 	if err := runCommand(serviceRoot, "swag", args...); err != nil {
// 		log.Printf("[auth-service] swag 命令执行失败，尝试使用 go run 回退: %v", err)
// 		fallbackArgs := append([]string{"run", "github.com/swaggo/swag/cmd/swag@latest"}, args...)
// 		if err := runCommand(serviceRoot, "go", fallbackArgs...); err != nil {
// 			return fmt.Errorf("使用 go run 更新 Swagger 失败: %w", err)
// 		}
// 	}

// 	log.Printf("[auth-service] Swagger 文档已同步")
// 	return nil
// }

// func runCommand(dir, name string, args ...string) error {
// 	cmd := exec.Command(name, args...)
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	cmd.Dir = dir
// 	return cmd.Run()
// }

// func resolveServiceRoot() (string, error) {
// 	_, filename, _, ok := runtime.Caller(0)
// 	if !ok {
// 		return "", fmt.Errorf("无法获取当前文件路径")
// 	}

// 	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..")), nil
// }
