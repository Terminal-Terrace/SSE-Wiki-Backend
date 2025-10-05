package database

import (
	"terminal-terrace/database"
	"time"

	"terminal-terrace/sse-wiki/config"
	"terminal-terrace/sse-wiki/internal/model"

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

	// 设置默认日志级别
	logLevel := databaseConf.LogLevel
	if logLevel == "" {
		logLevel = "info"
	}

	var err error
	PostgresDB, err = database.InitPostgres(
		&database.PostgresConfig{
			ServiceName:     "sse-wiki",
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
			Host:     redisConf.Host,
			Port:     redisConf.Port,
			Password: redisConf.Password,
			DB:       redisConf.DB,
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