package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hazuki-keatsu/bili-parse-api/models"
	"github.com/hazuki-keatsu/bili-parse-api/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController struct {
	db     *gorm.DB
	secret string
	encKey string
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Captcha  string `json:"captcha"`
}

type SessdataLoginRequest struct {
	SESSDATA   string `json:"sessdata" binding:"required"`
	BiliJCT    string `json:"bili_jct" binding:"required"`
	DedeUserID string `json:"dedeuserid" binding:"required"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Expires int64  `json:"expires,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewAuthController(db *gorm.DB, secret, encKey string) *AuthController {
	return &AuthController{
		db:     db,
		secret: secret,
		encKey: encKey,
	}
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "Invalid request parameters",
		})
		return
	}

	// 查找用户
	var user models.User
	if err := c.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// 生成JWT token
	token, expires, err := c.generateToken(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, LoginResponse{
			Success: false,
			Message: "Failed to generate token",
		})
		return
	}

	// 保存会话
	session := models.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Unix(expires, 0),
		IsValid:   true,
	}
	c.db.Create(&session)

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	c.db.Save(&user)

	ctx.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Token:   token,
		Expires: expires,
	})
}

func (c *AuthController) SessdataLogin(ctx *gin.Context) {
	var req SessdataLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "Invalid request parameters",
		})
		return
	}

	// 加密存储SESSDATA信息
	encSessdata, err := utils.Encrypt(req.SESSDATA, c.encKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, LoginResponse{
			Success: false,
			Message: "Encryption failed",
		})
		return
	}

	encBiliJCT, err := utils.Encrypt(req.BiliJCT, c.encKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, LoginResponse{
			Success: false,
			Message: "Encryption failed",
		})
		return
	}

	encDedeUserID, err := utils.Encrypt(req.DedeUserID, c.encKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, LoginResponse{
			Success: false,
			Message: "Encryption failed",
		})
		return
	}

	// 查找或创建用户
	var user models.User
	username := "sessdata_user_" + req.DedeUserID

	err = c.db.Where("username = ?", username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		// 创建新用户
		user = models.User{
			Username:   username,
			SESSDATA:   encSessdata,
			BiliJCT:    encBiliJCT,
			DedeUserID: encDedeUserID,
			IsActive:   true,
		}
		if err := c.db.Create(&user).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, LoginResponse{
				Success: false,
				Message: "Failed to create user",
			})
			return
		}
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, LoginResponse{
			Success: false,
			Message: "Database error",
		})
		return
	} else {
		// 更新用户信息
		user.SESSDATA = encSessdata
		user.BiliJCT = encBiliJCT
		user.DedeUserID = encDedeUserID
		c.db.Save(&user)
	}

	// 生成JWT token
	token, expires, err := c.generateToken(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, LoginResponse{
			Success: false,
			Message: "Failed to generate token",
		})
		return
	}

	// 保存会话
	session := models.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Unix(expires, 0),
		IsValid:   true,
	}
	c.db.Create(&session)

	ctx.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Token:   token,
		Expires: expires,
	})
}

func (c *AuthController) generateToken(userID uint) (string, int64, error) {
	expires := time.Now().Add(24 * time.Hour).Unix()

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expires,
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(c.secret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expires, nil
}
