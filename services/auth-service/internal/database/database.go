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

	var err error
	PostgresDB, err = database.InitPostgres(
		&database.PostgresConfig{
			Username:        databaseConf.Username,
			Password:        databaseConf.Password,
			Host:            databaseConf.Host,
			Port:            databaseConf.Port,
			Database:        databaseConf.Database,
			SSLMode:         databaseConf.SSLMode,
			MaxIdleConns:    databaseConf.MaxIdleConns,
			MaxOpenConns:    databaseConf.MaxOpenConns,
			ConnMaxLifetime: time.Duration(databaseConf.MaxLifetime) * time.Minute,
		},
	)

	if err != nil {
		panic(err)
	}

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
