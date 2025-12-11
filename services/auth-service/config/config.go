// config/config.go - 配置管理文件
package config

import (
	"fmt"
	"log"
	"strings"
	"sync"

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

// Load 加载配置文件
func Load(configPath string) error {
	var err error
	once.Do(func() {
		// 首先加载 .env 文件到环境变量
		envPath := "../../.env"
		if err = godotenv.Load(envPath); err != nil {
			log.Printf("警告: 无法加载 .env 文件: %v", err)
		}

		k = koanf.New(".")

		// 先加载配置文件
		if err = k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			err = fmt.Errorf("加载配置文件失败: %w", err)
			return
		}

		// 再加载环境变量（覆盖配置文件）
		if err = k.Load(env.Provider("", ".", func(s string) string {
			return strings.ReplaceAll(strings.ToLower(s), "_", ".")
		}), nil); err != nil {
			log.Printf("加载环境变量失败: %v", err)
		}

		// 解析到结构体
		Conf = &AppConfig{}
		if err = k.Unmarshal("", Conf); err != nil {
			err = fmt.Errorf("解析配置失败: %w", err)
			return
		}

		fmt.Println(Conf.Github)
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

	return nil
}
