package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"sync"
	"time"

	"terminal-terrace/sse-wiki/config"
	// 引入模型
	"terminal-terrace/sse-wiki/internal/model"
)

// var db *gorm.DB
var (
	db *gorm.DB
	mu sync.Mutex
)

func GetMysqlDb() *gorm.DB {
	mu.Lock()
	defer mu.Unlock()
	if db == nil {
		initPostgres()
	}
	return db
}

func initPostgres() {
	// 从环境变量获取数据库连接参数
	dbConfig := &config.Conf.Database
	username := dbConfig.Username
	password := dbConfig.Password
	host := dbConfig.Host
	port := dbConfig.Port
	dbname := dbConfig.Database

	// 设置默认值
	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 5432 // PostgreSQL默认端口
	}

	// 构建PostgreSQL DSN连接字符串
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		host, username, password, dbname, port)

	// 如果需要启用SSL，可以设置环境变量
	sslmode := dbConfig.SSLMode
	if sslmode {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=require TimeZone=Asia/Shanghai",
			host, username, password, dbname, port)
	}

	// 配置GORM
	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		Logger: logger.Default.LogMode(logger.Info), // 设置默认日志级别
	}

	// 根据环境调整日志级别
	logLevel := os.Getenv("DB_LOG_LEVEL")
	switch logLevel {
	case "silent":
		config.Logger = logger.Default.LogMode(logger.Silent)
	case "error":
		config.Logger = logger.Default.LogMode(logger.Error)
	case "warn":
		config.Logger = logger.Default.LogMode(logger.Warn)
	}

	// 连接数据库
	var err error
	db, err = gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期

	err = model.InitTable(db)
	if err != nil {
		log.Fatalf("Failed to initialize database tables: %v", err)
	}

	log.Println("Database connection established successfully")
}
