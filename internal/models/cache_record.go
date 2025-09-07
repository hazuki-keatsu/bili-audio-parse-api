package models

import (
	"time"
)

// CacheRecord 缓存记录模型
type CacheRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CacheKey  string    `gorm:"uniqueIndex;size:255" json:"cache_key"`
	BVID      string    `gorm:"index;size:20" json:"bvid"`
	Quality   int       `gorm:"index" json:"quality"`
	FilePath  string    `gorm:"size:500" json:"file_path"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
}

// RequestLog 请求日志模型
type RequestLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ClientIP    string    `gorm:"index;size:45" json:"client_ip"`
	UserAgent   string    `gorm:"size:500" json:"user_agent"`
	BVID        string    `gorm:"index;size:20" json:"bvid"`
	Quality     int       `json:"quality"`
	StatusCode  int       `gorm:"index" json:"status_code"`
	ErrorMsg    string    `gorm:"size:1000" json:"error_msg"`
	ProcessTime int64     `json:"process_time"` // 处理时间(毫秒)
	CreatedAt   time.Time `gorm:"index" json:"created_at"`
}

// AudioInfo 音频信息结构
type AudioInfo struct {
	URL         string `json:"url"`          // 本地音频文件链接
	OriginalURL string `json:"original_url"` // 原始B站链接
	Format      string `json:"format"`       // 格式 (mp3)
	Bitrate     int    `json:"bitrate"`      // 比特率
	Duration    int    `json:"duration"`     // 时长(秒)
	Quality     int    `json:"quality"`      // 音质编号
	Size        int64  `json:"size"`         // 文件大小
	FileName    string `json:"file_name"`    // 本地文件名
}
