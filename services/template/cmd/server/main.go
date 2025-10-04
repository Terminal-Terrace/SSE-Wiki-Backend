package main

import (
	"log"
	"time"

	"terminal-terrace/database"
	"terminal-terrace/template/config"
	"terminal-terrace/template/internal/model"
	"terminal-terrace/template/internal/route"
)

func main() {
	// 1. 加载配置
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化数据库连接
	dbConfig := &database.PostgresConfig{
		Username:        config.Conf.Database.Username,
		Password:        config.Conf.Database.Password,
		Host:            config.Conf.Database.Host,
		Port:            config.Conf.Database.Port,
		Database:        config.Conf.Database.Database,
		SSLMode:         config.Conf.Database.SSLMode,
		LogLevel:        config.Conf.Log.Level,
		MaxIdleConns:    config.Conf.Database.MaxIdleConns,
		MaxOpenConns:    config.Conf.Database.MaxOpenConns,
		ConnMaxLifetime: time.Duration(config.Conf.Database.MaxLifetime) * time.Second,
	}

	db, err := database.InitPostgres(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 3. 自动迁移数据库表
	if err := model.InitTable(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 4. 初始化路由（传入数据库连接）
	router := route.SetupRouter(db)

	// 5. 启动服务器
	addr := config.Conf.Server.Host + ":" + string(rune(config.Conf.Server.Port))
	log.Printf("Server starting on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
