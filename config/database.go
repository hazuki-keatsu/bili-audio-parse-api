package config

import (
	"github.com/glebarez/sqlite"
	"github.com/hazuki-keatsu/bili-parse-api/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDatabase(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移
	err = db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.ParseCache{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
