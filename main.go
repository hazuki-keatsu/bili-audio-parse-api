package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hazuki-keatsu/bili-parse-api/config"
	"github.com/hazuki-keatsu/bili-parse-api/routes"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 创建数据目录
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	if err := os.MkdirAll("./logs", 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// 初始化数据库
	db, err := config.InitDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 设置路由
	r := routes.SetupRoutes(db, cfg)

	// 启动服务器
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
