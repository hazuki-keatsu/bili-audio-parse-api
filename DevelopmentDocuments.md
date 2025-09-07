# B站DASH音频解析服务开发文档

## 一、项目概述

### 1.1 项目目标
本项目是一个基于Go语言开发的B站音频解析后端服务，使用Gin Web框架和GORM ORM框架，采用SQLite作为数据库。该服务为博客网站提供B站音频解析能力，支持通过BV号获取DASH格式音频资源。

### 1.2 技术栈
- **编程语言**：Go 1.25.1
- **Web框架**：Gin v1.10.1
- **ORM框架**：GORM v2.0+
- **数据库**：SQLite 3
- **其他依赖**：
  - `github.com/golang-jwt/jwt/v5` - JWT认证
  - `github.com/gin-contrib/cors` - 跨域支持
  - `golang.org/x/crypto` - 加密支持
  - `github.com/go-redis/redis/v8` - 缓存支持（可选）

### 1.3 项目结构
```
bili-parse-api/
├── main.go                 # 程序入口
├── config/                 # 配置管理
│   ├── config.go
│   └── database.go
├── models/                 # 数据模型
│   ├── user.go
│   ├── session.go
│   └── cache.go
├── controllers/            # 控制器
│   ├── auth.go
│   ├── parse.go
│   └── status.go
├── services/               # 业务逻辑
│   ├── bilibili.go
│   ├── auth.go
│   └── cache.go
├── middlewares/            # 中间件
│   ├── auth.go
│   ├── cors.go
│   └── ratelimit.go
├── utils/                  # 工具函数
│   ├── crypto.go
│   ├── http.go
│   └── wbi.go
├── routes/                 # 路由配置
│   └── routes.go
├── logs/                   # 日志文件
├── data/                   # 数据库文件
│   └── bili.db
└── README.md
```

## 二、环境配置与依赖管理

### 2.1 初始化项目
```bash
# 创建项目目录
mkdir bili-parse-api
cd bili-parse-api

# 初始化Go模块
go mod init github.com/hazuki-keatsu/bili-parse-api

# 安装依赖
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/sqlite
go get github.com/golang-jwt/jwt/v5
go get github.com/gin-contrib/cors
go get golang.org/x/crypto/bcrypt
go get github.com/sirupsen/logrus
```

### 2.2 配置文件设计
创建 `config/config.go`：
```go
package config

import (
    "os"
    "strconv"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Security SecurityConfig
    Cache    CacheConfig
    Bilibili BilibiliConfig
}

type ServerConfig struct {
    Host string
    Port string
    Mode string
}

type DatabaseConfig struct {
    Path string
}

type SecurityConfig struct {
    JWTSecret     string
    EncryptionKey string
}

type CacheConfig struct {
    TTL int // 缓存时间，单位：秒
}

type BilibiliConfig struct {
    UserAgent string
    Referer   string
}

func LoadConfig() *Config {
    return &Config{
        Server: ServerConfig{
            Host: getEnv("SERVER_HOST", "0.0.0.0"),
            Port: getEnv("SERVER_PORT", "8080"),
            Mode: getEnv("GIN_MODE", "debug"),
        },
        Database: DatabaseConfig{
            Path: getEnv("DB_PATH", "./data/bili.db"),
        },
        Security: SecurityConfig{
            JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
            EncryptionKey: getEnv("ENCRYPTION_KEY", "your-32-byte-key-here-123456789"),
        },
        Cache: CacheConfig{
            TTL: getEnvInt("CACHE_TTL", 3600), // 默认1小时
        },
        Bilibili: BilibiliConfig{
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            Referer:   "https://www.bilibili.com",
        },
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}
```

## 三、数据库设计

### 3.1 数据库连接
创建 `config/database.go`：
```go
package config

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func InitDatabase(dbPath string) (*gorm.DB, error) {
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, err
    }

    // 自动迁移
    err = db.AutoMigrate(
        &models.User{},
        &models.Session{},
        &models.ParseCache{},
    )
    if err != nil {
        return nil, err
    }

    return db, nil
}
```

### 3.2 数据模型设计
创建 `models/user.go`：
```go
package models

import (
    "time"
    "gorm.io/gorm"
)

// User 用户表，用于管理不同的B站账号
type User struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    Username    string    `gorm:"uniqueIndex;not null" json:"username"`
    Password    string    `gorm:"not null" json:"-"` // 加密存储
    SESSDATA    string    `gorm:"type:text" json:"-"` // 加密存储
    BiliJCT     string    `gorm:"type:text" json:"-"` // 加密存储
    DedeUserID  string    `gorm:"type:text" json:"-"` // 加密存储
    IsActive    bool      `gorm:"default:true" json:"is_active"`
    LastLoginAt *time.Time `json:"last_login_at"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
```

创建 `models/session.go`：
```go
package models

import (
    "time"
    "gorm.io/gorm"
)

// Session 会话表，用于管理登录状态
type Session struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"`
    Token     string    `gorm:"uniqueIndex;not null" json:"token"`
    ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
    IsValid   bool      `gorm:"default:true" json:"is_valid"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    
    User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
```

创建 `models/cache.go`：
```go
package models

import (
    "time"
    "gorm.io/gorm"
)

// ParseCache 解析缓存表
type ParseCache struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    BV        string    `gorm:"index;not null" json:"bv"`
    Quality   int       `gorm:"index" json:"quality"`
    UserID    uint      `gorm:"index" json:"user_id"` // 关联用户，因为不同用户权限不同
    AudioData string    `gorm:"type:text" json:"audio_data"` // JSON格式存储音频信息
    ExpiresAt time.Time `gorm:"index" json:"expires_at"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

## 四、核心服务实现

### 4.1 加密工具
创建 `utils/crypto.go`：
```go
package utils

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "io"
)

func Encrypt(plaintext, key string) (string, error) {
    block, err := aes.NewCipher([]byte(key))
    if err != nil {
        return "", err
    }

    plainTextBytes := []byte(plaintext)
    ciphertext := make([]byte, aes.BlockSize+len(plainTextBytes))
    iv := ciphertext[:aes.BlockSize]
    
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }

    stream := cipher.NewCFBEncrypter(block, iv)
    stream.XORKeyStream(ciphertext[aes.BlockSize:], plainTextBytes)

    return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(ciphertext, key string) (string, error) {
    ciphertextBytes, err := base64.URLEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher([]byte(key))
    if err != nil {
        return "", err
    }

    if len(ciphertextBytes) < aes.BlockSize {
        return "", errors.New("ciphertext too short")
    }

    iv := ciphertextBytes[:aes.BlockSize]
    ciphertextBytes = ciphertextBytes[aes.BlockSize:]

    stream := cipher.NewCFBDecrypter(block, iv)
    stream.XORKeyStream(ciphertextBytes, ciphertextBytes)

    return string(ciphertextBytes), nil
}
```

### 4.2 B站API服务
创建 `services/bilibili.go`：
```go
package services

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "time"
    
    "bili-parse-api/utils"
)

type BilibiliService struct {
    client *http.Client
}

type AudioInfo struct {
    URL      string `json:"url"`
    Format   string `json:"format"`
    Bitrate  int    `json:"bitrate"`
    Duration int    `json:"duration"`
    Quality  string `json:"quality"`
}

type BilibiliResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    struct {
        Dash struct {
            Audio []struct {
                ID       int    `json:"id"`
                BaseURL  string `json:"base_url"`
                Bitrate  int    `json:"bandwidth"`
                MimeType string `json:"mime_type"`
                Codecs   string `json:"codecs"`
            } `json:"audio"`
        } `json:"dash"`
        Duration int `json:"duration"`
    } `json:"data"`
}

func NewBilibiliService() *BilibiliService {
    return &BilibiliService{
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (s *BilibiliService) ParseAudio(bv string, quality int, cookies string) (*AudioInfo, error) {
    // 首先获取视频基本信息
    cid, err := s.getCID(bv)
    if err != nil {
        return nil, err
    }

    // 构建播放地址请求
    playURL := "https://api.bilibili.com/x/player/wbi/playurl"
    params := url.Values{
        "bvid":  {bv},
        "cid":   {strconv.Itoa(cid)},
        "fnval": {"4048"}, // DASH格式
        "fnver": {"0"},
        "fourk": {"1"},
    }

    // 添加WBI签名
    signedParams, err := utils.SignWBI(params)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("GET", playURL+"?"+signedParams, nil)
    if err != nil {
        return nil, err
    }

    // 设置请求头
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
    req.Header.Set("Referer", "https://www.bilibili.com/")
    req.Header.Set("Cookie", cookies)

    resp, err := s.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var biliResp BilibiliResponse
    if err := json.NewDecoder(resp.Body).Decode(&biliResp); err != nil {
        return nil, err
    }

    if biliResp.Code != 0 {
        return nil, fmt.Errorf("bilibili API error: %s", biliResp.Message)
    }

    // 选择合适的音频质量
    if len(biliResp.Data.Dash.Audio) == 0 {
        return nil, fmt.Errorf("no audio stream found")
    }

    audio := s.selectAudioByQuality(biliResp.Data.Dash.Audio, quality)
    
    return &AudioInfo{
        URL:      audio.BaseURL,
        Format:   s.parseFormat(audio.MimeType),
        Bitrate:  audio.Bitrate,
        Duration: biliResp.Data.Duration,
        Quality:  s.getQualityName(audio.ID),
    }, nil
}

func (s *BilibiliService) getCID(bv string) (int, error) {
    // 获取视频CID的实现
    apiURL := fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", bv)
    
    req, err := http.NewRequest("GET", apiURL, nil)
    if err != nil {
        return 0, err
    }

    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
    req.Header.Set("Referer", "https://www.bilibili.com/")

    resp, err := s.client.Do(req)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    var result struct {
        Code int `json:"code"`
        Data struct {
            CID int `json:"cid"`
        } `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return 0, err
    }

    if result.Code != 0 {
        return 0, fmt.Errorf("failed to get CID")
    }

    return result.Data.CID, nil
}

func (s *BilibiliService) selectAudioByQuality(audios []struct {
    ID       int    `json:"id"`
    BaseURL  string `json:"base_url"`
    Bitrate  int    `json:"bandwidth"`
    MimeType string `json:"mime_type"`
    Codecs   string `json:"codecs"`
}, quality int) struct {
    ID       int    `json:"id"`
    BaseURL  string `json:"base_url"`
    Bitrate  int    `json:"bandwidth"`
    MimeType string `json:"mime_type"`
    Codecs   string `json:"codecs"`
} {
    // 根据质量参数选择合适的音频流
    // 可以根据ID或bitrate进行选择
    if quality > 0 && quality < len(audios) {
        return audios[quality]
    }
    // 默认返回最高质量
    return audios[0]
}

func (s *BilibiliService) parseFormat(mimeType string) string {
    if strings.Contains(mimeType, "mp4") {
        return "m4s"
    }
    return "unknown"
}

func (s *BilibiliService) getQualityName(id int) string {
    qualityMap := map[int]string{
        30216: "64K",
        30232: "132K", 
        30280: "192K",
        30250: "Hi-Res无损",
    }
    if name, ok := qualityMap[id]; ok {
        return name
    }
    return "标准音质"
}
```

### 4.3 WBI签名工具
创建 `utils/wbi.go`：
```go
package utils

import (
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "net/url"
    "sort"
    "strconv"
    "strings"
    "time"
)

// WBI签名相关常量
const (
    mixinKeyEncTab = "fgjklmnopqrstuvwxyzabcdehi6789+/="
    wbiKey         = "7cd084941338484aae1ad9425b84077c4932caff0ff746eab6f01bf08b70ac45"
)

func SignWBI(params url.Values) (string, error) {
    // 添加时间戳
    params.Set("wts", strconv.FormatInt(time.Now().Unix(), 10))
    
    // 对参数进行排序
    keys := make([]string, 0, len(params))
    for k := range params {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    // 构建查询字符串
    var query strings.Builder
    for i, k := range keys {
        if i > 0 {
            query.WriteByte('&')
        }
        query.WriteString(url.QueryEscape(k))
        query.WriteByte('=')
        query.WriteString(url.QueryEscape(params.Get(k)))
    }

    // 计算签名
    hash := md5.Sum([]byte(query.String() + wbiKey))
    params.Set("w_rid", hex.EncodeToString(hash[:]))

    return params.Encode(), nil
}
```

## 五、API接口实现

### 5.1 认证控制器
创建 `controllers/auth.go`：
```go
package controllers

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
    
    "bili-parse-api/models"
    "bili-parse-api/utils"
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
```

### 5.2 解析控制器
创建 `controllers/parse.go`：
```go
package controllers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    "bili-parse-api/models"
    "bili-parse-api/services"
    "bili-parse-api/utils"
)

type ParseController struct {
    db         *gorm.DB
    biliSvc    *services.BilibiliService
    encKey     string
    cacheTTL   int
}

type ParseRequest struct {
    BV      string `form:"bv" binding:"required"`
    Quality int    `form:"quality"`
}

type ParseResponse struct {
    Success bool                   `json:"success"`
    Audio   *services.AudioInfo    `json:"audio,omitempty"`
    Message string                 `json:"message,omitempty"`
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
```

### 5.3 状态控制器
创建 `controllers/status.go`：
```go
package controllers

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    "bili-parse-api/models"
)

type StatusController struct {
    db *gorm.DB
}

type StatusResponse struct {
    Alive       bool                       `json:"alive"`
    LoginStatus map[string]string          `json:"login_status"`
    Statistics  map[string]interface{}     `json:"statistics"`
}

func NewStatusController(db *gorm.DB) *StatusController {
    return &StatusController{db: db}
}

func (c *StatusController) GetStatus(ctx *gin.Context) {
    // 检查数据库连接
    sqlDB, err := c.db.DB()
    alive := true
    if err != nil {
        alive = false
    } else if err := sqlDB.Ping(); err != nil {
        alive = false
    }

    // 获取登录状态
    loginStatus := c.getLoginStatus()
    
    // 获取统计信息
    statistics := c.getStatistics()

    ctx.JSON(http.StatusOK, StatusResponse{
        Alive:       alive,
        LoginStatus: loginStatus,
        Statistics:  statistics,
    })
}

func (c *StatusController) getLoginStatus() map[string]string {
    var sessions []models.Session
    c.db.Preload("User").Where("expires_at > ? AND is_valid = ?", time.Now(), true).Find(&sessions)

    status := make(map[string]string)
    for _, session := range sessions {
        if session.ExpiresAt.After(time.Now()) {
            status[session.User.Username] = "valid"
        } else {
            status[session.User.Username] = "expired"
        }
    }

    return status
}

func (c *StatusController) getStatistics() map[string]interface{} {
    stats := make(map[string]interface{})

    // 用户数量
    var userCount int64
    c.db.Model(&models.User{}).Count(&userCount)
    stats["total_users"] = userCount

    // 活跃会话数量
    var activeSessionCount int64
    c.db.Model(&models.Session{}).Where("expires_at > ? AND is_valid = ?", time.Now(), true).Count(&activeSessionCount)
    stats["active_sessions"] = activeSessionCount

    // 缓存条目数量
    var cacheCount int64
    c.db.Model(&models.ParseCache{}).Where("expires_at > ?", time.Now()).Count(&cacheCount)
    stats["cache_entries"] = cacheCount

    return stats
}
```

## 六、中间件实现

### 6.1 认证中间件
创建 `middlewares/auth.go`：
```go
package middlewares

import (
    "net/http"
    "strings"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "gorm.io/gorm"
    
    "bili-parse-api/models"
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
```

### 6.2 限流中间件
创建 `middlewares/ratelimit.go`：
```go
package middlewares

import (
    "net/http"
    "sync"
    "time"
    
    "github.com/gin-gonic/gin"
)

type RateLimit struct {
    requests map[string][]time.Time
    mu       sync.RWMutex
    limit    int
    window   time.Duration
}

func NewRateLimit(limit int, window time.Duration) *RateLimit {
    rl := &RateLimit{
        requests: make(map[string][]time.Time),
        limit:    limit,
        window:   window,
    }
    
    // 定期清理过期记录
    go rl.cleanup()
    
    return rl
}

func (rl *RateLimit) Middleware() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        clientIP := ctx.ClientIP()
        
        if !rl.allow(clientIP) {
            ctx.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": int(rl.window.Seconds()),
            })
            ctx.Abort()
            return
        }
        
        ctx.Next()
    }
}

func (rl *RateLimit) allow(key string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    now := time.Now()
    cutoff := now.Add(-rl.window)
    
    // 获取或创建请求记录
    requests := rl.requests[key]
    
    // 移除过期的请求记录
    validRequests := make([]time.Time, 0)
    for _, req := range requests {
        if req.After(cutoff) {
            validRequests = append(validRequests, req)
        }
    }
    
    // 检查是否超过限制
    if len(validRequests) >= rl.limit {
        rl.requests[key] = validRequests
        return false
    }
    
    // 添加当前请求
    validRequests = append(validRequests, now)
    rl.requests[key] = validRequests
    
    return true
}

func (rl *RateLimit) cleanup() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        rl.mu.Lock()
        cutoff := time.Now().Add(-rl.window)
        
        for key, requests := range rl.requests {
            validRequests := make([]time.Time, 0)
            for _, req := range requests {
                if req.After(cutoff) {
                    validRequests = append(validRequests, req)
                }
            }
            
            if len(validRequests) == 0 {
                delete(rl.requests, key)
            } else {
                rl.requests[key] = validRequests
            }
        }
        rl.mu.Unlock()
    }
}
```

## 七、路由配置

### 7.1 路由设置
创建 `routes/routes.go`：
```go
package routes

import (
    "time"
    
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    "bili-parse-api/config"
    "bili-parse-api/controllers"
    "bili-parse-api/middlewares"
    "bili-parse-api/services"
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
```

## 八、程序入口

### 8.1 主程序
创建 `main.go`：
```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "bili-parse-api/config"
    "bili-parse-api/models"
    "bili-parse-api/routes"
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

    // 自动迁移
    err = db.AutoMigrate(
        &models.User{},
        &models.Session{},
        &models.ParseCache{},
    )
    if err != nil {
        log.Fatalf("Failed to migrate database: %v", err)
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
```

## 九、部署与运行

### 9.1 构建项目
```bash
# 安装依赖
go mod tidy

# 构建可执行文件
go build -o bili-parse-api main.go

# Windows构建
GOOS=windows GOARCH=amd64 go build -o bili-parse-api.exe main.go

# Linux构建  
GOOS=linux GOARCH=amd64 go build -o bili-parse-api main.go
```

### 9.2 环境变量配置
创建 `.env` 文件：
```env
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
GIN_MODE=release
DB_PATH=./data/bili.db
JWT_SECRET=your-secret-key-here
ENCRYPTION_KEY=your-32-byte-key-here-123456789
CACHE_TTL=3600
```

### 9.3 运行服务
```bash
# 直接运行
./bili-parse-api

# 或使用systemd (Linux)
# 创建服务文件 /etc/systemd/system/bili-parse-api.service
```

### 9.4 Docker部署
创建 `Dockerfile`：
```dockerfile
FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bili-parse-api main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/bili-parse-api .
RUN mkdir -p data logs

EXPOSE 8080
CMD ["./bili-parse-api"]
```

创建 `docker-compose.yml`：
```yaml
version: '3.8'
services:
  bili-parse-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=release
      - JWT_SECRET=your-secret-key
      - ENCRYPTION_KEY=your-32-byte-key-here-123456789
    volumes:
      - ./data:/root/data
      - ./logs:/root/logs
    restart: unless-stopped
```

## 十、API使用示例

### 10.1 登录示例
```bash
# 账号密码登录
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"your_username","password":"your_password"}'

# SESSDATA登录
curl -X POST http://localhost:8080/api/login/sessdata \
  -H "Content-Type: application/json" \
  -d '{"sessdata":"your_sessdata","bili_jct":"your_bili_jct","dedeuserid":"your_dedeuserid"}'
```

### 10.2 解析音频示例
```bash
curl -X GET "http://localhost:8080/api/parse?bv=BV1234567890&quality=0" \
  -H "Authorization: Bearer your_jwt_token"
```

### 10.3 状态查询示例
```bash
curl -X GET http://localhost:8080/api/status
```

## 十一、测试与维护

### 11.1 单元测试
创建测试文件，例如 `controllers/auth_test.go`：
```go
package controllers

import (
    "testing"
    // 添加测试相关的导入
)

func TestLogin(t *testing.T) {
    // 实现登录测试
}

func TestSessdataLogin(t *testing.T) {
    // 实现SESSDATA登录测试
}
```

### 11.2 日志管理
建议集成结构化日志，如 logrus：
```go
import "github.com/sirupsen/logrus"

// 在main.go中配置日志
func setupLogger() {
    logrus.SetFormatter(&logrus.JSONFormatter{})
    logrus.SetLevel(logrus.InfoLevel)
    
    file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err == nil {
        logrus.SetOutput(file)
    }
}
```

### 11.3 监控与告警
- 集成 Prometheus 指标收集
- 配置健康检查端点
- 设置数据库连接监控
- 配置API响应时间监控

## 十二、注意事项

1. **安全性**：
   - 定期更新JWT密钥
   - 使用HTTPS部署
   - 限制API访问频率
   - 加密存储敏感信息

2. **合规性**：
   - 遵守B站服务条款
   - 尊重版权保护
   - 避免过度请求

3. **维护性**：
   - 定期清理过期缓存
   - 监控服务状态
   - 备份重要数据
   - 更新B站接口适配

4. **性能优化**：
   - 合理设置缓存时间
   - 数据库索引优化
   - 连接池配置
   - 并发请求控制

这个开发文档提供了完整的B站音频解析服务实现方案，基于Go语言、Gin框架、GORM和SQLite数据库，满足了需求文档中的所有功能要求。
