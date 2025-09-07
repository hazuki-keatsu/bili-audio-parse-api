package utils

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Code:    statusCode,
		Message: message,
	})
}

// IsValidBVID 验证BV号格式
func IsValidBVID(bvid string) bool {
	// BV号格式：BV + 10位数字字母组合
	matched, _ := regexp.MatchString(`^BV[a-zA-Z0-9]{10}$`, bvid)
	return matched
}
