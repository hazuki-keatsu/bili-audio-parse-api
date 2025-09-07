package config

import (
	"fmt"
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
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
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
			// 配置文件未找到，使用默认值
			fmt.Println("Config file not found, using defaults")
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
}
