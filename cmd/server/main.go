package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"bili-parse-api/internal/api/routes"
	"bili-parse-api/internal/core/cache"
	"bili-parse-api/internal/core/config"
	"bili-parse-api/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 初始化配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 根据配置创建必要的目录和文件
	if err := createRequiredDirectories(cfg); err != nil {
		log.Fatal("Failed to create required directories:", err)
	}

	// 初始化日志系统
	if err := initLogging(cfg); err != nil {
		log.Fatal("Failed to initialize logging:", err)
	}

	// 初始化数据库
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 数据库迁移
	if err := autoMigrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 初始化并启动缓存清理工作器
	cacheManager := cache.NewManager(cfg.Cache.Dir, cfg.Cache.TTL, db)
	cacheManager.StartCleanupWorker(cfg.Cache.CleanupInterval)
	logrus.Infof("Cache cleanup worker started with interval: %v", cfg.Cache.CleanupInterval)

	// 启动日志清理任务
	startLogCleanup(cfg)
	if cfg.Logging.File != "" && cfg.Logging.MaxAge > 0 {
		logrus.Infof("Log cleanup worker started, will clean logs older than %d days", cfg.Logging.MaxAge)
	}

	// 初始化路由
	router := routes.SetupRouter(cfg, db)

	// 启动服务器
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	logrus.Infof("Server starting on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// 配置GORM日志级别，避免"record not found"错误日志
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), // 只记录警告和错误，不记录info级别的"record not found"
	}

	switch cfg.Database.Type {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.Database.DSN), gormConfig)
	default:
		// 可以在这里添加MySQL等其他数据库支持
		db, err = gorm.Open(sqlite.Open("data.db"), gormConfig)
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.CacheRecord{},
		&models.RequestLog{},
	)
}

// createRequiredDirectories 根据配置创建程序运行所需的目录和文件
func createRequiredDirectories(cfg *config.Config) error {
	// 1. 创建缓存目录
	if cfg.Cache.Dir != "" {
		if err := os.MkdirAll(cfg.Cache.Dir, 0755); err != nil {
			return fmt.Errorf("failed to create cache directory %s: %w", cfg.Cache.Dir, err)
		}
		logrus.Infof("Cache directory created/verified: %s", cfg.Cache.Dir)
	}

	// 2. 创建日志目录和文件
	if cfg.Logging.File != "" {
		logDir := filepath.Dir(cfg.Logging.File)
		if logDir != "." && logDir != "" {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return fmt.Errorf("failed to create log directory %s: %w", logDir, err)
			}
			logrus.Infof("Log directory created/verified: %s", logDir)
		}

		// 确保日志文件存在（如果不存在则创建空文件）
		if _, err := os.Stat(cfg.Logging.File); os.IsNotExist(err) {
			if file, err := os.Create(cfg.Logging.File); err != nil {
				return fmt.Errorf("failed to create log file %s: %w", cfg.Logging.File, err)
			} else {
				file.Close()
				logrus.Infof("Log file created: %s", cfg.Logging.File)
			}
		}
	}

	// 3. 确保数据库文件的目录存在（对于SQLite）
	if cfg.Database.Type == "sqlite" && cfg.Database.DSN != "" {
		dbDir := filepath.Dir(cfg.Database.DSN)
		if dbDir != "." && dbDir != "" {
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return fmt.Errorf("failed to create database directory %s: %w", dbDir, err)
			}
			logrus.Infof("Database directory created/verified: %s", dbDir)
		}

		// 检查数据库文件是否存在，如果不存在会由GORM自动创建
		if _, err := os.Stat(cfg.Database.DSN); os.IsNotExist(err) {
			logrus.Infof("Database file will be created: %s", cfg.Database.DSN)
		}
	}

	return nil
}

// initLogging 初始化日志系统
func initLogging(cfg *config.Config) error {
	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		return fmt.Errorf("invalid log level %s: %w", cfg.Logging.Level, err)
	}
	logrus.SetLevel(level)

	// 设置日志格式
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 如果配置了日志文件，设置文件输出
	if cfg.Logging.File != "" {
		// 打开日志文件
		logFile, err := os.OpenFile(cfg.Logging.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %w", cfg.Logging.File, err)
		}

		// 同时输出到控制台和文件
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		logrus.SetOutput(multiWriter)

		logrus.Infof("Logging initialized: level=%s, file=%s", cfg.Logging.Level, cfg.Logging.File)
	} else {
		// 只输出到控制台
		logrus.SetOutput(os.Stdout)
		logrus.Infof("Logging initialized: level=%s, output=stdout", cfg.Logging.Level)
	}

	return nil
}

// startLogCleanup 启动日志清理任务
func startLogCleanup(cfg *config.Config) {
	if cfg.Logging.File == "" || cfg.Logging.MaxAge <= 0 {
		return
	}

	// 使用配置的清理间隔，如果未配置则默认24小时
	interval := cfg.Logging.CleanupInterval
	if interval == 0 {
		interval = 24 * time.Hour
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// 启动时先执行一次清理
		cleanupOldLogs(cfg)

		for range ticker.C {
			cleanupOldLogs(cfg)
		}
	}()
}

// cleanupOldLogs 清理过期日志文件
func cleanupOldLogs(cfg *config.Config) {
	if cfg.Logging.File == "" {
		return
	}

	logDir := filepath.Dir(cfg.Logging.File)
	baseFileName := filepath.Base(cfg.Logging.File)

	// 获取文件名前缀（不含扩展名）
	ext := filepath.Ext(baseFileName)
	prefix := baseFileName[:len(baseFileName)-len(ext)]

	// 扫描日志目录
	files, err := filepath.Glob(filepath.Join(logDir, prefix+"*"))
	if err != nil {
		logrus.Warnf("Failed to scan log directory %s: %v", logDir, err)
		return
	}

	cutoff := time.Now().AddDate(0, 0, -cfg.Logging.MaxAge)
	cleanedCount := 0

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		// 删除超过MaxAge天的文件
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(file); err != nil {
				logrus.Warnf("Failed to remove old log file %s: %v", file, err)
			} else {
				cleanedCount++
			}
		}
	}

	if cleanedCount > 0 {
		logrus.Infof("Cleaned up %d old log files", cleanedCount)
	}
}
