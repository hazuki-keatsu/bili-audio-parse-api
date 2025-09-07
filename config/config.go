package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Security SecurityConfig
	Cache    CacheConfig
	Bilibili BilibiliConfig
}

type ServerConfig struct {
	Host string
	Port string
	Mode string
}

type DatabaseConfig struct {
	Path string
}

type SecurityConfig struct {
	JWTSecret     string
	EncryptionKey string
}

type CacheConfig struct {
	TTL int // 缓存时间，单位：秒
}

type BilibiliConfig struct {
	UserAgent string
	Referer   string
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./data/bili.db"),
		},
		Security: SecurityConfig{
			JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
			EncryptionKey: getEnv("ENCRYPTION_KEY", "your-32-byte-key-here-123456789"),
		},
		Cache: CacheConfig{
			TTL: getEnvInt("CACHE_TTL", 3600), // 默认1小时
		},
		Bilibili: BilibiliConfig{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			Referer:   "https://www.bilibili.com",
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
