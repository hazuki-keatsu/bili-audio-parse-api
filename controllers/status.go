package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hazuki-keatsu/bili-parse-api/models"
	"gorm.io/gorm"
)

type StatusController struct {
	db *gorm.DB
}

type StatusResponse struct {
	Alive       bool                   `json:"alive"`
	LoginStatus map[string]string      `json:"login_status"`
	Statistics  map[string]interface{} `json:"statistics"`
}

func NewStatusController(db *gorm.DB) *StatusController {
	return &StatusController{db: db}
}

func (c *StatusController) GetStatus(ctx *gin.Context) {
	// 检查数据库连接
	sqlDB, err := c.db.DB()
	alive := true
	if err != nil {
		alive = false
	} else if err := sqlDB.Ping(); err != nil {
		alive = false
	}

	// 获取登录状态
	loginStatus := c.getLoginStatus()

	// 获取统计信息
	statistics := c.getStatistics()

	ctx.JSON(http.StatusOK, StatusResponse{
		Alive:       alive,
		LoginStatus: loginStatus,
		Statistics:  statistics,
	})
}

func (c *StatusController) getLoginStatus() map[string]string {
	var sessions []models.Session
	c.db.Preload("User").Where("expires_at > ? AND is_valid = ?", time.Now(), true).Find(&sessions)

	status := make(map[string]string)
	for _, session := range sessions {
		if session.ExpiresAt.After(time.Now()) {
			status[session.User.Username] = "valid"
		} else {
			status[session.User.Username] = "expired"
		}
	}

	return status
}

func (c *StatusController) getStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// 用户数量
	var userCount int64
	c.db.Model(&models.User{}).Count(&userCount)
	stats["total_users"] = userCount

	// 活跃会话数量
	var activeSessionCount int64
	c.db.Model(&models.Session{}).Where("expires_at > ? AND is_valid = ?", time.Now(), true).Count(&activeSessionCount)
	stats["active_sessions"] = activeSessionCount

	// 缓存条目数量
	var cacheCount int64
	c.db.Model(&models.ParseCache{}).Where("expires_at > ?", time.Now()).Count(&cacheCount)
	stats["cache_entries"] = cacheCount

	return stats
}
