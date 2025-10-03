package database

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type PostgresConfig struct {
	Username        string
	Password        string
	Host            string
	Port            int
	Database        string
	SSLMode         bool
	LogLevel        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

func InitPostgres(dbConfig *PostgresConfig) (*gorm.DB, error) {
	if dbConfig == nil {
		// 从环境变量获取配置
		return nil, fmt.Errorf("dbConfig cannot be nil")
	}
	if port, err := strconv.Atoi(os.Getenv("DB_PORT")); err == nil {
		dbConfig.Port = port
	}
	if sslmode := os.Getenv("DB_SSLMODE"); sslmode == "true" {
		dbConfig.SSLMode = true
	}
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
	switch dbConfig.LogLevel {
	case "silent":
		config.Logger = logger.Default.LogMode(logger.Silent)
	case "error":
		config.Logger = logger.Default.LogMode(logger.Error)
	case "warn":
		config.Logger = logger.Default.LogMode(logger.Warn)
	default:
		config.Logger = logger.Default.LogMode(logger.Info)
	}

	// 连接数据库
	var err error
	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sqlDB from gorm DB: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)       // 最大空闲连接数
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)       // 最大打开连接数
	sqlDB.SetConnMaxLifetime(dbConfig.ConnMaxLifetime) // 连接最大生命周期

	return db, nil
}
