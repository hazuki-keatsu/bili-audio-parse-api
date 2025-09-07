package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 记录HTTP请求日志
		logrus.WithFields(logrus.Fields{
			"type":       "http_request",
			"client_ip":  param.ClientIP,
			"method":     param.Method,
			"path":       param.Path,
			"status":     param.StatusCode,
			"latency":    param.Latency.String(),
			"user_agent": param.Request.UserAgent(),
			"timestamp":  param.TimeStamp.Format(time.RFC3339),
		}).Info("HTTP Request")

		return ""
	})
}

// LogError 记录错误日志
func LogError(err error, context map[string]interface{}) {
	fields := logrus.Fields{
		"type":  "error",
		"error": err.Error(),
	}

	// 添加上下文信息
	for k, v := range context {
		fields[k] = v
	}

	logrus.WithFields(fields).Error("Application Error")
}

// LogInfo 记录信息日志
func LogInfo(message string, context map[string]interface{}) {
	fields := logrus.Fields{
		"type":    "info",
		"message": message,
	}

	// 添加上下文信息
	for k, v := range context {
		fields[k] = v
	}

	logrus.WithFields(fields).Info(message)
}
