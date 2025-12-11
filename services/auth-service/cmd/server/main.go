package main

import (
	"fmt"
	"log"

	"terminal-terrace/auth-service/config"
	"terminal-terrace/auth-service/internal/database"
	grpcserver "terminal-terrace/auth-service/internal/grpc"
	"terminal-terrace/auth-service/internal/model"
	"terminal-terrace/email"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. 加载配置
	config.MustLoad("config.yaml")

	// 2. 确保数据库存在
	if err := ensureDatabaseExists(); err != nil {
		log.Fatalf("[auth-service] 数据库创建失败: %v", err)
	}

	// 3. 初始化数据库连接
	database.InitDatabase()

	// 4. 同步数据库结构
	if err := model.InitTable(database.GetDB()); err != nil {
		log.Fatalf("[auth-service] 数据库表初始化失败: %v", err)
	}

	// 5. 初始化邮件客户端（gRPC 服务需要）
	mailer := email.NewClient(&config.Conf.Smtp)

	// 6. 启动 gRPC server (blocking)
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
