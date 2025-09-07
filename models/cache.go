package models

import (
	"time"

	"gorm.io/gorm"
)

// ParseCache 解析缓存表
type ParseCache struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	BV        string         `gorm:"index;not null" json:"bv"`
	Quality   int            `gorm:"index" json:"quality"`
	UserID    uint           `gorm:"index" json:"user_id"`        // 关联用户，因为不同用户权限不同
	AudioData string         `gorm:"type:text" json:"audio_data"` // JSON格式存储音频信息
	ExpiresAt time.Time      `gorm:"index" json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
