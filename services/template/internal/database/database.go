package database

import (
	"terminal-terrace/database"
	"time"

	"terminal-terrace/template/config"
	"terminal-terrace/template/internal/model"

	"gorm.io/gorm"
)

var (
	PostgresDB *gorm.DB
	RedisDB    *database.RedisClient
)

func InitDatabase() {
	initPostgres()
	initRedis()
}

func initPostgres() {
	databaseConf := config.Conf.Database

	logLevel := databaseConf.LogLevel
	if logLevel == "" {
		logLevel := "info"
	}

	var err error
	PostgresDB, err = database.InitPostgres(
		&database.PostgresConfig{
			ServiceName:     "template-service",
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

	// 初始化数据库表
	err = model.InitTable(PostgresDB)
	if err != nil {
		panic(err)
	}
}

func initRedis() {
	redisConf := config.Conf.Redis

	var err error
	RedisDB, err = database.InitRedis(
		&database.RedisConfig{
			ServiceName: "template-service",
			Host:        redisConf.Host,
			Port:        redisConf.Port,
			Password:    redisConf.Password,
			DB:          redisConf.DB,
			PoolSize:    redisConf.PoolSize,
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

// GetRedis 获取 Redis 实例
func GetRedis() *database.RedisClient {
	return RedisDB
}
