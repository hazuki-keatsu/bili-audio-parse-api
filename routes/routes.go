package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hazuki-keatsu/bili-parse-api/config"
	"github.com/hazuki-keatsu/bili-parse-api/controllers"
	"github.com/hazuki-keatsu/bili-parse-api/middlewares"
	"github.com/hazuki-keatsu/bili-parse-api/services"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB, cfg *config.Config) *gin.Engine {
	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	r := gin.Default()

	// CORS中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 限流中间件
	rateLimit := middlewares.NewRateLimit(20, time.Minute) // 每分钟20次请求
	r.Use(rateLimit.Middleware())

	// 初始化服务
	biliSvc := services.NewBilibiliService()

	// 初始化控制器
	authCtrl := controllers.NewAuthController(db, cfg.Security.JWTSecret, cfg.Security.EncryptionKey)
	parseCtrl := controllers.NewParseController(db, biliSvc, cfg.Security.EncryptionKey, cfg.Cache.TTL)
	statusCtrl := controllers.NewStatusController(db)

	// API路由组
	api := r.Group("/api")
	{
		// 认证路由（无需token）
		api.POST("/login", authCtrl.Login)
		api.POST("/login/sessdata", authCtrl.SessdataLogin)

		// 状态查询（无需token）
		api.GET("/status", statusCtrl.GetStatus)

		// 需要认证的路由
		protected := api.Group("")
		protected.Use(middlewares.AuthMiddleware(db, cfg.Security.JWTSecret))
		{
			protected.GET("/parse", parseCtrl.ParseAudio)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
