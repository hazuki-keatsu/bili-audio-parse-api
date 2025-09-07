package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hazuki-keatsu/bili-parse-api/models"
	"github.com/hazuki-keatsu/bili-parse-api/services"
	"github.com/hazuki-keatsu/bili-parse-api/utils"
	"gorm.io/gorm"
)

type ParseController struct {
	db       *gorm.DB
	biliSvc  *services.BilibiliService
	encKey   string
	cacheTTL int
}

type ParseRequest struct {
	BV      string `form:"bv" binding:"required"`
	Quality int    `form:"quality"`
}

type ParseResponse struct {
	Success bool                `json:"success"`
	Audio   *services.AudioInfo `json:"audio,omitempty"`
	Message string              `json:"message,omitempty"`
}

func NewParseController(db *gorm.DB, biliSvc *services.BilibiliService, encKey string, cacheTTL int) *ParseController {
	return &ParseController{
		db:       db,
		biliSvc:  biliSvc,
		encKey:   encKey,
		cacheTTL: cacheTTL,
	}
}

func (c *ParseController) ParseAudio(ctx *gin.Context) {
	var req ParseRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ParseResponse{
			Success: false,
			Message: "Invalid request parameters",
		})
		return
	}

	// 从JWT获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ParseResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	uid := userID.(uint)

	// 检查缓存
	if audio := c.getFromCache(req.BV, req.Quality, uid); audio != nil {
		ctx.JSON(http.StatusOK, ParseResponse{
			Success: true,
			Audio:   audio,
		})
		return
	}

	// 获取用户的Cookie信息
	cookies, err := c.getUserCookies(uid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ParseResponse{
			Success: false,
			Message: "Failed to get user credentials",
		})
		return
	}

	// 调用B站API解析音频
	audio, err := c.biliSvc.ParseAudio(req.BV, req.Quality, cookies)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ParseResponse{
			Success: false,
			Message: fmt.Sprintf("Parse failed: %v", err),
		})
		return
	}

	// 保存到缓存
	c.saveToCache(req.BV, req.Quality, uid, audio)

	ctx.JSON(http.StatusOK, ParseResponse{
		Success: true,
		Audio:   audio,
	})
}

func (c *ParseController) getFromCache(bv string, quality int, userID uint) *services.AudioInfo {
	var cache models.ParseCache
	err := c.db.Where("bv = ? AND quality = ? AND user_id = ? AND expires_at > ?",
		bv, quality, userID, time.Now()).First(&cache).Error

	if err != nil {
		return nil
	}

	var audio services.AudioInfo
	if err := json.Unmarshal([]byte(cache.AudioData), &audio); err != nil {
		return nil
	}

	return &audio
}

func (c *ParseController) saveToCache(bv string, quality int, userID uint, audio *services.AudioInfo) {
	audioData, err := json.Marshal(audio)
	if err != nil {
		return
	}

	cache := models.ParseCache{
		BV:        bv,
		Quality:   quality,
		UserID:    userID,
		AudioData: string(audioData),
		ExpiresAt: time.Now().Add(time.Duration(c.cacheTTL) * time.Second),
	}

	c.db.Create(&cache)
}

func (c *ParseController) getUserCookies(userID uint) (string, error) {
	var user models.User
	if err := c.db.First(&user, userID).Error; err != nil {
		return "", err
	}

	sessdata, err := utils.Decrypt(user.SESSDATA, c.encKey)
	if err != nil {
		return "", err
	}

	biliJCT, err := utils.Decrypt(user.BiliJCT, c.encKey)
	if err != nil {
		return "", err
	}

	dedeUserID, err := utils.Decrypt(user.DedeUserID, c.encKey)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("SESSDATA=%s; bili_jct=%s; DedeUserID=%s",
		sessdata, biliJCT, dedeUserID), nil
}
