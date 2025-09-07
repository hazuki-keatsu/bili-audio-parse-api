package cache

import (
	"bili-parse-api/internal/models"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// Manager 缓存管理器
type Manager struct {
	cacheDir string
	ttl      time.Duration
	db       *gorm.DB
}

// CacheItem 缓存项
type CacheItem struct {
	Key       string            `json:"key"`
	Data      *models.AudioInfo `json:"data"`
	CreatedAt time.Time         `json:"created_at"`
	ExpiresAt time.Time         `json:"expires_at"`
}

// NewManager 创建缓存管理器
func NewManager(cacheDir string, ttl time.Duration, db *gorm.DB) *Manager {
	// 确保缓存目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		// 记录错误但不中断程序
		fmt.Printf("Warning: failed to create cache directory: %v\n", err)
	}

	return &Manager{
		cacheDir: cacheDir,
		ttl:      ttl,
		db:       db,
	}
}

// Get 获取缓存
func (m *Manager) Get(bvid string, quality int) *models.AudioInfo {
	key := m.generateKey(bvid, quality)

	// 1. 检查数据库记录
	var record models.CacheRecord
	if err := m.db.Where("cache_key = ? AND expires_at > ?", key, time.Now()).First(&record).Error; err != nil {
		return nil
	}

	// 2. 读取缓存文件
	filePath := filepath.Join(m.cacheDir, key+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 缓存文件不存在，清理数据库记录
		m.db.Delete(&record)
		return nil
	}

	var item CacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		// 缓存文件损坏，清理
		os.Remove(filePath)
		m.db.Delete(&record)
		return nil
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		// 过期，清理
		os.Remove(filePath)
		// 如果有MP3文件也要清理
		if item.Data != nil && item.Data.FileName != "" {
			mp3Path := filepath.Join(m.cacheDir, item.Data.FileName)
			os.Remove(mp3Path)
		}
		m.db.Delete(&record)
		return nil
	}

	// 检查MP3文件是否存在
	if item.Data != nil && item.Data.FileName != "" {
		mp3Path := filepath.Join(m.cacheDir, item.Data.FileName)
		if _, err := os.Stat(mp3Path); err != nil {
			// MP3文件不存在，清理缓存
			os.Remove(filePath)
			m.db.Delete(&record)
			return nil
		}
	}

	return item.Data
}

// Set 设置缓存
func (m *Manager) Set(bvid string, quality int, audioInfo *models.AudioInfo) error {
	key := m.generateKey(bvid, quality)

	item := CacheItem{
		Key:       key,
		Data:      audioInfo,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(m.ttl),
	}

	// 1. 写入缓存文件
	filePath := filepath.Join(m.cacheDir, key+".json")
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal cache item: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	// 2. 记录到数据库
	record := models.CacheRecord{
		CacheKey:  key,
		BVID:      bvid,
		Quality:   quality,
		FilePath:  filePath,
		ExpiresAt: item.ExpiresAt,
	}

	// 使用 Create 或 Save 来插入或更新记录
	if err := m.db.Where(models.CacheRecord{CacheKey: key}).Assign(record).FirstOrCreate(&record).Error; err != nil {
		// 数据库写入失败，但不删除缓存文件，因为缓存仍然有效
		fmt.Printf("Warning: failed to save cache record to database: %v\n", err)
	}

	return nil
}

// CleanupExpired 清理过期缓存
func (m *Manager) CleanupExpired() error {
	// 1. 从数据库中找到过期记录
	var expiredRecords []models.CacheRecord
	if err := m.db.Where("expires_at < ?", time.Now()).Find(&expiredRecords).Error; err != nil {
		return fmt.Errorf("failed to find expired records: %w", err)
	}

	// 2. 删除文件和数据库记录
	for _, record := range expiredRecords {
		// 删除文件
		if err := os.Remove(record.FilePath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to remove cache file %s: %v\n", record.FilePath, err)
		}

		// 删除数据库记录
		if err := m.db.Delete(&record).Error; err != nil {
			fmt.Printf("Warning: failed to delete cache record %d: %v\n", record.ID, err)
		}
	}

	return nil
}

// generateKey 生成缓存键
func (m *Manager) generateKey(bvid string, quality int) string {
	data := bvid + "_" + strconv.Itoa(quality)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// StartCleanupWorker 启动清理工作协程
func (m *Manager) StartCleanupWorker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := m.CleanupExpired(); err != nil {
				fmt.Printf("Cache cleanup error: %v\n", err)
			}
		}
	}()
}
