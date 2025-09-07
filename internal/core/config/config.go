package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Cache     CacheConfig     `mapstructure:"cache"`
	Bilibili  BilibiliConfig  `mapstructure:"bilibili"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	CORS      CORSConfig      `mapstructure:"cors"`
	Logging   LoggingConfig   `mapstructure:"logging"`
}

type ServerConfig struct {
	Host  string `mapstructure:"host"`
	Port  string `mapstructure:"port"`
	Debug bool   `mapstructure:"debug"`
}

type DatabaseConfig struct {
	Type string `mapstructure:"type"`
	DSN  string `mapstructure:"dsn"`
}

type CacheConfig struct {
	Dir             string        `mapstructure:"dir"`
	TTL             time.Duration `mapstructure:"ttl"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

type BilibiliConfig struct {
	UserAgent string        `mapstructure:"user_agent"`
	Referer   string        `mapstructure:"referer"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
	Burst             int  `mapstructure:"burst"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

type LoggingConfig struct {
	Level           string        `mapstructure:"level"`
	File            string        `mapstructure:"file"`
	MaxSize         int           `mapstructure:"max_size"`
	MaxBackups      int           `mapstructure:"max_backups"`
	MaxAge          int           `mapstructure:"max_age"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件未找到，但不自动创建，使用默认值
			fmt.Println("Config file not found, using default values. You can create configs/config.yaml manually.")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.debug", false)

	// Database defaults
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.dsn", "data.db")

	// Cache defaults
	viper.SetDefault("cache.dir", "./parse_cache")
	viper.SetDefault("cache.ttl", "1h")
	viper.SetDefault("cache.cleanup_interval", "30m")

	// Bilibili defaults
	viper.SetDefault("bilibili.user_agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	viper.SetDefault("bilibili.referer", "https://www.bilibili.com")
	viper.SetDefault("bilibili.timeout", "30s")

	// Rate limit defaults
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.requests_per_minute", 20)
	viper.SetDefault("rate_limit.burst", 5)

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", []string{"*"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"*"})

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "./logs/app.log")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 7)
	viper.SetDefault("logging.max_age", 30)
	viper.SetDefault("logging.cleanup_interval", "24h")
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig() error {
	// 确保configs目录存在
	configDir := "./configs"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 默认配置内容
	defaultConfig := `server:
  host: "0.0.0.0"
  port: "8080"
  debug: false

database:
  type: "sqlite"
  dsn: "./data.db"

cache:
  dir: "./parse_cache"
  ttl: "24h"
  cleanup_interval: "1h"

bilibili:
  user_agent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
  referer: "https://www.bilibili.com"
  timeout: "30s"

rate_limit:
  enabled: true
  requests_per_minute: 20
  burst: 5

cors:
  allowed_origins: ["*"]
  allowed_methods: ["GET", "POST", "OPTIONS"]
  allowed_headers: ["*"]

logging:
  level: "info"
  file: "./logs/app.log"
  max_size: 100
  max_backups: 7
  max_age: 30
  cleanup_interval: "24h"
`

	configFile := filepath.Join(configDir, "config.yaml")
	return os.WriteFile(configFile, []byte(defaultConfig), 0644)
}
