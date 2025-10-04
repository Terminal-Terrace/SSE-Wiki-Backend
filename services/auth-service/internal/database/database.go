package database

import (
	"terminal-terrace/database"
	"time"

	"terminal-terrace/auth-service/config"

	"gorm.io/gorm"
)

var (
	PostgresDB *gorm.DB
	RedisDB    *database.RedisClient
)

func InitDatabase() {
	databaseConf := config.Conf.Database
	redisConf := config.Conf.Redis

	logLevel := databaseConf.LogLevel
	if logLevel == "" {
		logLevel = "silent"
	}

	var err error
	PostgresDB, err = database.InitPostgres(
		&database.PostgresConfig{
			ServiceName:     "auth-service",
			Username:        databaseConf.Username,
			Password:        databaseConf.Password,
			Host:            databaseConf.Host,
			Port:            databaseConf.Port,
			Database:        databaseConf.Database,
			SSLMode:         databaseConf.SSLMode,
			LogLevel:        logLevel,
			MaxIdleConns:    databaseConf.MaxIdleConns,
			MaxOpenConns:    databaseConf.MaxOpenConns,
			ConnMaxLifetime: time.Duration(databaseConf.MaxLifetime) * time.Second,
		},
	)

	if err != nil {
		panic(err)
	}

	// 初始化 Redis
	RedisDB, err = database.InitRedis(
		&database.RedisConfig{
			ServiceName: "auth-service",
			Host:        redisConf.Host,
			Port:        redisConf.Port,
			Password:    redisConf.Password,
			DB:          redisConf.DB,
		},
	)

	if err != nil {
		panic(err)
	}
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return PostgresDB
}
