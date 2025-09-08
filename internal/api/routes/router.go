package routes

import (
	"github.com/hazuki-keatsu/bili-audio-parse-api/internal/api/handlers"
	"github.com/hazuki-keatsu/bili-audio-parse-api/internal/api/middleware"
	"github.com/hazuki-keatsu/bili-audio-parse-api/internal/core/cache"
	"github.com/hazuki-keatsu/bili-audio-parse-api/internal/core/config"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter 设置路由
func SetupRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	// 设置Gin模式
	if !cfg.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// 基础中间件
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	// CORS中间件
	router.Use(middleware.CORS(
		cfg.CORS.AllowedOrigins,
		cfg.CORS.AllowedMethods,
		cfg.CORS.AllowedHeaders,
	))

	// 限流中间件
	if cfg.RateLimit.Enabled {
		router.Use(middleware.RateLimit(
			cfg.RateLimit.RequestsPerMinute,
			time.Minute,
		))
	}

	// 初始化缓存管理器
	cacheManager := cache.NewManager(cfg.Cache.Dir, cfg.Cache.TTL, db)

	// 启动缓存清理任务
	cacheManager.StartCleanupWorker(cfg.Cache.CleanupInterval)

	// 初始化处理器
	parseHandler := handlers.NewParseHandler(
		cfg.Bilibili.UserAgent,
		cfg.Bilibili.Referer,
		cfg.Cache.Dir,
		cacheManager,
		db,
	)
	statusHandler := handlers.NewStatusHandler(db)

	// 静态文件服务器 - 提供MP3文件访问
	router.Static("/static", cfg.Cache.Dir)

	// API路由组
	v1 := router.Group("/api/v1")
	{
		v1.GET("/parse", parseHandler.ParseAudio)    // 音频解析
		v1.GET("/status", statusHandler.GetStatus)   // 服务状态
		v1.GET("/health", statusHandler.HealthCheck) // 健康检查
	}

	return router
}
