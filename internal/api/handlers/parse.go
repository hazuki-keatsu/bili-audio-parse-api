package handlers

import (
	"bili-parse-api/internal/core/bilibili"
	"bili-parse-api/internal/core/cache"
	"bili-parse-api/internal/models"
	"bili-parse-api/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ParseHandler struct {
	parser *bilibili.AudioParser
	cache  *cache.Manager
	db     *gorm.DB
}

func NewParseHandler(userAgent, referer, cacheDir string, cacheManager *cache.Manager, db *gorm.DB) *ParseHandler {
	return &ParseHandler{
		parser: bilibili.NewAudioParser(userAgent, referer, cacheDir),
		cache:  cacheManager,
		db:     db,
	}
}

// ParseRequest 解析请求结构
type ParseRequest struct {
	BV      string `form:"bv" binding:"required" json:"bv"` // BV号
	Quality int    `form:"quality" json:"quality"`          // 音质 (可选)
	Token   string `form:"token" json:"token"`              // 访问令牌 (可选)
}

// ParseResponse 解析响应结构
type ParseResponse struct {
	Success bool              `json:"success"`
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    *models.AudioInfo `json:"data,omitempty"`
}

// ParseAudio 解析音频接口
func (h *ParseHandler) ParseAudio(c *gin.Context) {
	startTime := time.Now()

	var req ParseRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logRequest(c, req.BV, req.Quality, http.StatusBadRequest, err.Error(), startTime)
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 验证BV号格式
	if !utils.IsValidBVID(req.BV) {
		h.logRequest(c, req.BV, req.Quality, http.StatusBadRequest, "无效的BV号格式", startTime)
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的BV号格式")
		return
	}

	// 1. 检查缓存
	if cached := h.cache.Get(req.BV, req.Quality); cached != nil {
		h.logRequest(c, req.BV, req.Quality, http.StatusOK, "", startTime)
		utils.SuccessResponse(c, cached)
		return
	}

	// 2. 解析音频
	audioInfo, err := h.parser.ParseAudio(req.BV, req.Quality)
	if err != nil {
		h.logRequest(c, req.BV, req.Quality, http.StatusInternalServerError, err.Error(), startTime)
		utils.ErrorResponse(c, http.StatusInternalServerError, "解析失败: "+err.Error())
		return
	}

	// 3. 缓存结果
	if err := h.cache.Set(req.BV, req.Quality, audioInfo); err != nil {
		// 缓存失败不影响正常响应，只记录警告
		// 可以考虑添加日志记录
	}

	h.logRequest(c, req.BV, req.Quality, http.StatusOK, "", startTime)
	utils.SuccessResponse(c, audioInfo)
}

// logRequest 记录请求日志
func (h *ParseHandler) logRequest(c *gin.Context, bvid string, quality, statusCode int, errorMsg string, startTime time.Time) {
	processTime := time.Since(startTime).Milliseconds()

	log := models.RequestLog{
		ClientIP:    c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
		BVID:        bvid,
		Quality:     quality,
		StatusCode:  statusCode,
		ErrorMsg:    errorMsg,
		ProcessTime: processTime,
	}

	// 异步记录日志，不影响响应
	go func() {
		if err := h.db.Create(&log).Error; err != nil {
			// 日志记录失败，可以考虑使用其他日志系统
		}
	}()
}
