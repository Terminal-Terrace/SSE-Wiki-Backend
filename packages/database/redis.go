package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis 配置
type RedisConfig struct {
	ServiceName  string        // 服务名称，用于日志标识
	Host         string        // Redis 地址
	Port         int           // Redis 端口
	Password     string        // Redis 密码
	DB           int           // Redis 数据库编号
	PoolSize     int           // 连接池大小
	MinIdleConns int           // 最小空闲连接数
	MaxConnAge   time.Duration // 连接最大生命周期
}

// RedisClient Redis 客户端封装
type RedisClient struct {
	*redis.Client
}

// InitRedis 初始化 Redis 连接
func InitRedis(config *RedisConfig) (*RedisClient, error) {
	if config == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 设置默认值
	setRedisDefaults(config)

	// 创建 Redis 客户端
	options := &redis.Options{
		Addr:            fmt.Sprintf("%s:%d", config.Host, config.Port),
		DB:              config.DB,
		PoolSize:        config.PoolSize,
		MinIdleConns:    config.MinIdleConns,
		ConnMaxLifetime: config.MaxConnAge,
	}
	
	// 只有当密码不为空时才设置密码
	if config.Password != "" {
		log.Printf("[%s] 设置Redis密码", config.ServiceName)
		options.Password = config.Password
	} else {
		log.Printf("[%s] Redis无密码连接", config.ServiceName)
	}
	
	client := redis.NewClient(options)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %v", err)
	}

	// 简洁的连接成功日志
	serviceName := config.ServiceName
	if serviceName == "" {
		serviceName = "unknown-service"
	}
	log.Printf("[%s] Redis连接成功", serviceName)

	return &RedisClient{Client: client}, nil
}

// setRedisDefaults 设置默认值
func setRedisDefaults(c *RedisConfig) {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 6379
	}
	if c.PoolSize == 0 {
		c.PoolSize = 10
	}
	if c.MinIdleConns == 0 {
		c.MinIdleConns = 5
	}
	if c.MaxConnAge == 0 {
		c.MaxConnAge = 1 * time.Hour
	}
}
