package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresConfig PostgreSQL 配置
type PostgresConfig struct {
	ServiceName     string        // 服务名称，用于日志标识
	Username        string        // 数据库用户名
	Password        string        // 数据库密码
	Host            string        // 数据库地址
	Port            int           // 数据库端口
	Database        string        // 数据库名称
	SSLMode         bool          // 是否启用 SSL
	LogLevel        string        // 日志级别: silent, error, warn, info
	MaxIdleConns    int           // 最大空闲连接数
	MaxOpenConns    int           // 最大打开连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
}

// InitPostgres 初始化 PostgreSQL 连接
func InitPostgres(config *PostgresConfig) (*gorm.DB, error) {
	if config == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 设置默认值
	setDefaults(config)

	// 构建 DSN
	dsn := buildDSN(config)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: getLogger(config.LogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %v", err)
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// 简洁的连接成功日志
	serviceName := config.ServiceName
	if serviceName == "" {
		serviceName = "unknown-service"
	}
	log.Printf("[%s] 数据库连接成功", serviceName)
	return db, nil
}

// setDefaults 设置默认值
func setDefaults(c *PostgresConfig) {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 5432
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 10
	}
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 100
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = 1 * time.Hour
	}
}

// buildDSN 构建连接字符串
func buildDSN(c *PostgresConfig) string {
	sslmode := "disable"
	if c.SSLMode {
		sslmode = "require"
	}
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		c.Host, c.Username, c.Password, c.Database, c.Port, sslmode)
}

// getLogger 获取日志配置
func getLogger(level string) logger.Interface {
	switch level {
	case "silent":
		return logger.Default.LogMode(logger.Silent)
	case "error":
		return logger.Default.LogMode(logger.Error)
	case "warn":
		return logger.Default.LogMode(logger.Warn)
	case "info":
		return logger.Default.LogMode(logger.Info)
	default:
		return logger.Default.LogMode(logger.Info)
	}
}
