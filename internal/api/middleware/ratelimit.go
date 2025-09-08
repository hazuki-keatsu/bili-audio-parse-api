package middleware

import (
	"github.com/hazuki-keatsu/bili-audio-parse-api/internal/utils"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	requests int
	window   time.Duration
}

type visitor struct {
	tokens   int
	lastSeen time.Time
}

// NewRateLimiter 创建限流器
func NewRateLimiter(requests int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		requests: requests,
		window:   window,
	}

	// 启动清理协程
	go rl.cleanupVisitors()

	return rl
}

// RateLimit 限流中间件
func RateLimit(requests int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(requests, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.allow(ip) {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			tokens:   rl.requests - 1,
			lastSeen: now,
		}
		return true
	}

	// 重置窗口
	if now.Sub(v.lastSeen) > rl.window {
		v.tokens = rl.requests
		v.lastSeen = now
	}

	if v.tokens > 0 {
		v.tokens--
		v.lastSeen = now
		return true
	}

	return false
}

func (rl *rateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()

		for ip, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}

		rl.mu.Unlock()
	}
}
