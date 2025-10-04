// config/config.go - 配置管理文件
// AI一键生成的, 之后大概率要改
package config

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	Conf *AppConfig
	once sync.Once
	k    *koanf.Koanf
)

// AppConfig 应用配置结构
type AppConfig struct {
	Server   ServerConfig   `koanf:"server"`
	Database DatabaseConfig `koanf:"database"`
	Redis    RedisConfig    `koanf:"redis"`
	Log      LogConfig      `koanf:"log"`
	JWT      JWTConfig      `koanf:"jwt"`
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
	LogLevel     string `koanf:"log_level"`    // 数据库日志级别
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

// Load 加载配置文件
func Load(configPath string) error {
	var err error
	once.Do(func() {
		// 首先加载 .env 文件到环境变量
		if err = godotenv.Load("../../.env"); err != nil {
			log.Printf("警告: 无法加载 .env 文件: %v", err)
		}

		k = koanf.New(".")

		// 加载配置文件
		if err = k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			err = fmt.Errorf("加载配置文件失败: %w", err)
			return
		}

		// 加载环境变量（会覆盖配置文件）
		if err = k.Load(env.Provider("", ".", func(s string) string {
			return strings.Replace(strings.ToLower(s), "_", ".", -1)
		}), nil); err != nil {
			log.Printf("加载环境变量失败: %v", err)
		}

		// 解析到结构体
		Conf = &AppConfig{}
		if err = k.Unmarshal("", Conf); err != nil {
			err = fmt.Errorf("解析配置失败: %w", err)
			return
		}

		// 转换时间单位
		Conf.Server.ReadTimeout = Conf.Server.ReadTimeout * time.Second
		Conf.Server.WriteTimeout = Conf.Server.WriteTimeout * time.Second
	})

	return err
}

// MustLoad 加载配置，失败则 panic
func MustLoad(configPath string) {
	if err := Load(configPath); err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
}

// GetString 获取字符串配置
func GetString(key string) string {
	if k == nil {
		log.Fatal("配置未初始化")
	}
	return k.String(key)
}

// GetInt 获取整数配置
func GetInt(key string) int {
	if k == nil {
		log.Fatal("配置未初始化")
	}
	return k.Int(key)
}

// GetBool 获取布尔配置
func GetBool(key string) bool {
	if k == nil {
		log.Fatal("配置未初始化")
	}
	return k.Bool(key)
}

// Reload 重新加载配置
func Reload(configPath string) error {
	if k == nil {
		return fmt.Errorf("配置未初始化")
	}

	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return err
	}

	Conf = &AppConfig{}
	if err := k.Unmarshal("", Conf); err != nil {
		return err
	}

	Conf.Server.ReadTimeout = Conf.Server.ReadTimeout * time.Second
	Conf.Server.WriteTimeout = Conf.Server.WriteTimeout * time.Second

	return nil
}
