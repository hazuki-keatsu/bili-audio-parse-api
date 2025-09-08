package handlers

import (
	"github.com/hazuki-keatsu/bili-audio-parse-api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StatusHandler struct {
	db *gorm.DB
}

func NewStatusHandler(db *gorm.DB) *StatusHandler {
	return &StatusHandler{
		db: db,
	}
}

// StatusResponse 状态响应结构
type StatusResponse struct {
	Alive       bool                   `json:"alive"`
	LoginStatus map[string]string      `json:"loginStatus"`
	Stats       map[string]interface{} `json:"stats"`
}

// GetStatus 获取服务状态
func (h *StatusHandler) GetStatus(c *gin.Context) {
	stats := make(map[string]interface{})

	// 获取缓存统计
	var cacheCount int64
	if err := h.db.Model(&struct {
		ID uint `gorm:"primaryKey"`
	}{}).Count(&cacheCount).Error; err == nil {
		stats["cache_count"] = cacheCount
	}

	// 获取请求统计
	var requestCount int64
	if err := h.db.Model(&struct {
		ID uint `gorm:"primaryKey"`
	}{}).Count(&requestCount).Error; err == nil {
		stats["total_requests"] = requestCount
	}

	response := StatusResponse{
		Alive: true,
		LoginStatus: map[string]string{
			"anonymous": "valid", // 当前只支持匿名访问
		},
		Stats: stats,
	}

	utils.SuccessResponse(c, response)
}

// HealthCheck 健康检查
func (h *StatusHandler) HealthCheck(c *gin.Context) {
	// 检查数据库连接
	sqlDB, err := h.db.DB()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Database connection failed")
		return
	}

	if err := sqlDB.Ping(); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Database ping failed")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"status": "healthy",
		"timestamp": gin.H{
			"unix": gin.H{},
		},
	})
}
