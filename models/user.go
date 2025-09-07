package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表，用于管理不同的B站账号
type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Username    string         `gorm:"uniqueIndex;not null" json:"username"`
	Password    string         `gorm:"not null" json:"-"`  // 加密存储
	SESSDATA    string         `gorm:"type:text" json:"-"` // 加密存储
	BiliJCT     string         `gorm:"type:text" json:"-"` // 加密存储
	DedeUserID  string         `gorm:"type:text" json:"-"` // 加密存储
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	LastLoginAt *time.Time     `json:"last_login_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
