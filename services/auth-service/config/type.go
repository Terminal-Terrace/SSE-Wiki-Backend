package config

import (
	"terminal-terrace/email"
	"time"
)

// AppConfig 应用配置结构
type AppConfig struct {
	Server   ServerConfig      `koanf:"server"`
	GRPC     GRPCConfig        `koanf:"grpc"`
	Database DatabaseConfig    `koanf:"database"`
	Redis    RedisConfig       `koanf:"redis"`
	Log      LogConfig         `koanf:"log"`
	JWT      JWTConfig         `koanf:"jwt"`
	Smtp     email.Config      `koanf:"smtp"`
	Github   GithubOAuthConfig `koanf:"github"`
}

type GRPCConfig struct {
	Port int `koanf:"port"`
}

type ServerConfig struct {
	Host         string        `koanf:"host"`
	Port         int           `koanf:"port"`
	Mode         string        `koanf:"mode"` // debug, release
	ReadTimeout  time.Duration `koanf:"read_timeout"`
	WriteTimeout time.Duration `koanf:"write_timeout"`
}

type DatabaseConfig struct {
	Driver       string `koanf:"driver"`
	Host         string `koanf:"host"`
	Port         int    `koanf:"port"`
	Username     string `koanf:"username"`
	Password     string `koanf:"password"`
	Database     string `koanf:"database"`
	SSLMode      bool   `koanf:"sslmode"`
	LogLevel     string `koanf:"log_level"` // 数据库日志级别
	MaxOpenConns int    `koanf:"max_open_conns"`
	MaxIdleConns int    `koanf:"max_idle_conns"`
	MaxLifetime  int    `koanf:"max_lifetime"` // 秒
}

type RedisConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Password string `koanf:"password"`
	DB       int    `koanf:"db"`
	PoolSize int    `koanf:"pool_size"`
}

type LogConfig struct {
	Level  string `koanf:"level"`  // debug, info, warn, error
	Format string `koanf:"format"` // json, text
	Output string `koanf:"output"` // stdout, file
	Path   string `koanf:"path"`   // 日志文件路径
}

type JWTConfig struct {
	Secret     string `koanf:"secret"`
	ExpireTime int    `koanf:"expire_time"` // 小时
}

type GithubOAuthConfig struct {
	ClientID     string `koanf:"id"`
	ClientSecret string `koanf:"secret"`
}
