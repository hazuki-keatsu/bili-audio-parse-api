package main

import (
	"log"

	"bili-parse-api/internal/api/routes"
	"bili-parse-api/internal/core/config"
	"bili-parse-api/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 初始化配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
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

	switch cfg.Database.Type {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
	default:
		// 可以在这里添加MySQL等其他数据库支持
		db, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
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
