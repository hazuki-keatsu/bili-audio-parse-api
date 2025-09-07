package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hazuki-keatsu/bili-parse-api/models"
	"gorm.io/gorm"
)

func AuthMiddleware(db *gorm.DB, secret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			ctx.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			ctx.Abort()
			return
		}

		userID := uint(claims["user_id"].(float64))

		// 验证会话是否存在且有效
		var session models.Session
		err = db.Where("user_id = ? AND token = ? AND expires_at > ? AND is_valid = ?",
			userID, tokenString, time.Now(), true).First(&session).Error

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired or invalid"})
			ctx.Abort()
			return
		}

		ctx.Set("user_id", userID)
		ctx.Set("token", tokenString)
		ctx.Next()
	}
}
