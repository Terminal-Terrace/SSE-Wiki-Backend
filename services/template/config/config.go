// config/config.go - 配置管理文件
package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

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
		k = koanf.New(".")

		// 1. 加载配置文件
		if err = k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			err = fmt.Errorf("加载配置文件失败: %w", err)
			return
		}

		// 2. 加载标准环境变量（APP_ 前缀）
		// 例如：APP_DATABASE_HOST -> database.host
		if err = k.Load(env.Provider("APP_", ".", func(s string) string {
			return strings.Replace(strings.ToLower(
				strings.TrimPrefix(s, "APP_")), "_", ".", -1)
		}), nil); err != nil {
			log.Printf("加载环境变量失败: %v", err)
		}

		// 3. 加载简化的环境变量名（向后兼容）
		loadCustomEnvVars(k)

		// 4. 解析到结构体
		Conf = &AppConfig{}
		if err = k.Unmarshal("", Conf); err != nil {
			err = fmt.Errorf("解析配置失败: %w", err)
			return
		}

		// 5. 转换时间单位
		Conf.Server.ReadTimeout = Conf.Server.ReadTimeout * time.Second
		Conf.Server.WriteTimeout = Conf.Server.WriteTimeout * time.Second

		// 6. 验证必需配置
		if err = validateConfig(Conf); err != nil {
			return
		}
	})

	return err
}

// loadCustomEnvVars 加载自定义环境变量名（简化命名）
func loadCustomEnvVars(k *koanf.Koanf) {
	// 数据库配置
	if v := os.Getenv("DB_HOST"); v != "" {
		k.Set("database.host", v)
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		k.Set("database.port", v)
	}
	if v := os.Getenv("DB_USERNAME"); v != "" {
		k.Set("database.username", v)
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		k.Set("database.password", v)
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		k.Set("database.sslmode", v == "true")
	}

	// Redis 配置
	if v := os.Getenv("REDIS_HOST"); v != "" {
		k.Set("redis.host", v)
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		k.Set("redis.port", v)
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		k.Set("redis.password", v)
	}

	// JWT 配置
	if v := os.Getenv("JWT_SECRET"); v != "" {
		k.Set("jwt.secret", v)
	}
	if v := os.Getenv("JWT_EXPIRE_TIME"); v != "" {
		k.Set("jwt.expire_time", v)
	}

	// 日志级别
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		k.Set("log.level", v)
	}

	// 前端 URL（用于 CORS）
	if v := os.Getenv("FRONTEND_URL"); v != "" {
		k.Set("frontend_url", v)
	}
}

// validateConfig 验证配置的有效性
func validateConfig(conf *AppConfig) error {
	// 验证必需的配置
	if conf.Database.Password == "" {
		log.Println("⚠️  Warning: database.password is empty, please set DB_PASSWORD environment variable")
	}

	if conf.JWT.Secret == "" {
		log.Println("⚠️  Warning: jwt.secret is empty, please set JWT_SECRET environment variable")
	}

	return nil
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
