package testutils

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"terminal-terrace/sse-wiki/internal/model"
	dbPkg "terminal-terrace/database"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates a test database connection using environment variables
// Defaults to test database configuration if env vars not set
// Automatically migrates all tables before returning the connection
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Get database connection string from environment or use defaults
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		host := getEnvOrDefault("POSTGRES_HOST", "localhost")
		port := getEnvOrDefault("POSTGRES_PORT", "5433")
		user := getEnvOrDefault("POSTGRES_USER", "test")
		password := getEnvOrDefault("POSTGRES_PASSWORD", "test")
		dbname := getEnvOrDefault("POSTGRES_DB", "sse_wiki_test")
		
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Suppress logs in tests
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Initialize all tables
	if err := model.InitTable(db); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Return a transaction for automatic rollback
	tx := db.Begin()
	t.Cleanup(func() {
		tx.Rollback()
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	return tx
}

// SetupTestRedis creates a test Redis connection
// Returns nil if Redis is not available (tests can skip Redis-dependent features)
func SetupTestRedis(t *testing.T) *dbPkg.RedisClient {
	t.Helper()

	redisHost := getEnvOrDefault("REDIS_HOST", "localhost")
	redisPortStr := getEnvOrDefault("REDIS_PORT", "6380")
	redisPort, err := strconv.Atoi(redisPortStr)
	if err != nil || redisPort == 0 {
		redisPort = 6380
	}

	// Try to initialize Redis, but don't fail if it's not available
	redisClient, err := dbPkg.InitRedis(&dbPkg.RedisConfig{
		ServiceName: "sse-wiki-test",
		Host:        redisHost,
		Port:        redisPort,
		Password:    "",
		DB:          0,
	})
	if err == nil && redisClient != nil {
		// Cleanup: flush Redis on test cleanup
		t.Cleanup(func() {
			redisClient.FlushDB(context.Background())
		})
		return redisClient
	}

	// Redis not available, return nil (tests can skip)
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

