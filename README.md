# B站DASH音频解析服务

一个基于Go + Gin + GORM的B站DASH音频解析服务，支持通过BV号获取音频资源并自动下载转换为MP3格式提供本地访问。

## 功能特点

- 🎵 **音频解析**: 支持B站BV号解析，获取DASH格式音频流
- 🔄 **自动下载**: 自动下载m4s音频文件并转换为MP3格式
- 🛠️ **格式转换**: 使用FFmpeg将音频转换为通用的MP3格式
- 🔗 **本地服务**: 返回服务器本地MP3文件链接，避免防盗链问题
- 🔄 **WBI签名**: 实现B站WBI签名算法，确保请求合法性
- 💾 **智能缓存**: 本地文件缓存 + 数据库记录，提升响应速度
- 🛡️ **安全防护**: IP限流、参数验证、错误处理
- 📊 **状态监控**: 服务状态查询、健康检查接口
- 🌐 **跨域支持**: 配置灵活的CORS策略

## 快速开始

### 环境要求

- Go 1.21+
- FFmpeg (用于音频格式转换)
- SQLite (默认) 或 MySQL

### 安装和运行

1. **克隆项目**
```bash
git clone https://github.com/hazuki-keatsu/bili-audio-parse-api.git
cd bili-audio-parse-api
```

2. **安装依赖**
```bash
go mod tidy
```

3. **安装FFmpeg** (如果未安装)
```bash
# Ubuntu/Debian
sudo apt update && sudo apt install ffmpeg

# CentOS/RHEL
sudo yum install epel-release && sudo yum install ffmpeg

# macOS
brew install ffmpeg

# 验证安装
ffmpeg -version
```

4. **创建配置文件**
```bash
cp configs/config.example.yaml configs/config.yaml
```

5. **创建必要目录和文件**
```bash
mkdir -p parse_cache logs
touch data.db
```

6. **运行服务**
```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动

### Docker部署

1. **使用Docker Compose (推荐)**
```bash
docker-compose up -d
```

2. **或使用Docker**
```bash
docker build -t bili-parse-api .
docker run -p 8080:8080 -v $(pwd)/parse_cache:/root/parse_cache bili-parse-api
```

## API接口

### 解析音频

**GET** `/api/v1/parse`

**参数:**
- `bv` (必须): B站视频BV号，如 `BV1xx411c7mD`
- `quality` (可选): 音质代码，如 `30280` (192K)，默认为最高音质

**响应示例:**
```json
{
  "success": true,
  "code": 0,
  "message": "success",
  "data": {
    "url": "/static/BV1xx411c7mD_30280_1694123456.mp3",
    "original_url": "https://xy.mcdn.bilivideo.cn/path/to/audio.m4s",
    "format": "mp3",
    "bitrate": 192,
    "duration": 180,
    "quality": 30280,
    "size": 4096000,
    "file_name": "BV1xx411c7mD_30280_1694123456.mp3"
  }
}
```

**说明:**
- `url`: 本地MP3文件的访问路径，可直接用于播放
- `original_url`: 原始B站音频链接
- `file_name`: 本地缓存的文件名

### 服务状态

**GET** `/api/v1/status`

**响应示例:**
```json
{
  "success": true,
  "code": 0,
  "message": "success", 
  "data": {
    "alive": true,
    "loginStatus": {
      "anonymous": "valid"
    },
    "stats": {
      "cache_count": 1250,
      "total_requests": 5680
    }
  }
}
```

### 健康检查

**GET** `/api/v1/health`

**响应示例:**
```json
{
  "success": true,
  "code": 0,
  "message": "success",
  "data": {
    "status": "healthy"
  }
}
```

## 使用示例

### curl命令
```bash
# 解析音频
curl "http://localhost:8080/api/v1/parse?bv=BV1xx411c7mD"

# 指定音质解析
curl "http://localhost:8080/api/v1/parse?bv=BV1xx411c7mD&quality=30280"

# 检查服务状态  
curl "http://localhost:8080/api/v1/status"

# 检查健康状态  
curl "http://localhost:8080/api/v1/health"

# 直接下载MP3文件
curl -o "audio.mp3" "http://localhost:8080/static/BV1xx411c7mD_30280_1694123456.mp3"
```

### JavaScript
```javascript
// 解析音频
fetch('http://localhost:8080/api/v1/parse/audio?bvid=BV1xx411c7mD')
  .then(response => response.json())
  .then(data => {
    if (data.success) {
      console.log('本地MP3链接:', data.data.url);
      // 直接使用本地链接播放，无需处理防盗链
      const audio = new Audio('http://localhost:8080' + data.data.url);
      audio.play();
      
      // 或者提供下载链接
      const downloadLink = document.createElement('a');
      downloadLink.href = 'http://localhost:8080' + data.data.url;
      downloadLink.download = data.data.file_name;
      downloadLink.click();
    }
  });
```

### Python
```python
import requests
import os

# 解析音频
response = requests.get('http://localhost:8080/api/v1/parse/audio', params={
    'bvid': 'BV1xx411c7mD',
    'quality': 30280
})

data = response.json()
if data['success']:
    local_url = data['data']['url']
    file_name = data['data']['file_name']
    
    print(f"本地MP3链接: http://localhost:8080{local_url}")
    
    # 下载MP3文件到本地
    mp3_response = requests.get(f"http://localhost:8080{local_url}")
    with open(file_name, 'wb') as f:
        f.write(mp3_response.content)
    print(f"音频已保存为: {file_name}")
```

## 配置说明

配置文件位于 `configs/config.yaml`，主要配置项：

```yaml
server:
  host: "0.0.0.0"     # 监听地址
  port: "8080"        # 监听端口
  debug: false        # 调试模式

cache:
  dir: "./parse_cache"      # 缓存目录
  ttl: "24h"               # 缓存过期时间
  cleanup_interval: "1h"    # 清理间隔

rate_limit:
  enabled: true             # 是否启用限流
  requests_per_minute: 20   # 每分钟请求限制

ffmpeg:
  path: "ffmpeg"           # FFmpeg可执行文件路径 (可选)
  timeout: "5m"            # 转换超时时间
```

## 工作原理

1. **接收请求**: 用户通过API传入BV号
2. **获取信息**: 使用WBI签名算法获取视频信息和播放地址
3. **解析音频**: 从DASH格式中提取最佳音质的音频流URL
4. **下载转换**: 下载m4s格式音频文件并使用FFmpeg转换为MP3
5. **缓存存储**: 将转换后的MP3文件存储在本地缓存目录
6. **返回链接**: 返回本地MP3文件的访问链接

## 优势特点

✅ **无防盗链问题**: 音频文件存储在本地服务器，避免B站防盗链限制  
✅ **格式通用性**: 转换为MP3格式，兼容性更好  
✅ **访问速度快**: 本地文件访问，响应速度快  
✅ **支持下载**: 可直接下载MP3文件到本地  
✅ **缓存机制**: 相同内容无需重复下载转换

## 音质代码

| 代码 | 音质 | 说明 |
|------|------|------|
| 30216 | 64K | 低音质 |
| 30232 | 132K | 中音质 |
| 30280 | 192K | 高音质 |
| 30250 | 杜比全景声 | 需要大会员 |
| 30251 | Hi-Res无损 | 需要大会员 |

## 注意事项

⚠️ **重要提醒**

1. **仅供学习**: 本项目仅用于技术学习和研究，请遵守B站用户协议
2. **存储空间**: 音频文件会占用本地存储空间，建议定期清理缓存
3. **FFmpeg依赖**: 需要系统安装FFmpeg，确保转换功能正常
4. **转换时间**: 首次请求需要下载转换时间，后续访问缓存文件速度快
5. **权限限制**: 某些音频需要登录或大会员权限才能获取
6. **请求频率**: 请合理控制请求频率，避免触发B站风控

## 故障排除

### 常见问题

1. **获取音频失败**
   - 检查BV号格式是否正确
   - 确认视频是否存在且可访问
   - 检查网络连接

2. **FFmpeg转换失败**
   - 确认FFmpeg已正确安装: `ffmpeg -version`
   - 检查FFmpeg路径配置
   - 查看日志文件获取详细错误信息

3. **文件下载失败**
   - 检查网络连接稳定性
   - 确认parse_cache目录有写入权限
   - 检查磁盘剩余空间

4. **请求被限流**
   - 降低请求频率
   - 检查IP是否被B站风控

5. **缓存目录权限错误**
   - 确保应用有写入权限: `chmod 755 parse_cache`
   - 检查磁盘空间是否充足

## 开发指南

### 项目结构
```
bili-parse-api/
├── cmd/server/          # 程序入口
├── internal/
│   ├── api/            # API层
│   │   ├── handlers/   # 请求处理器
│   │   ├── middleware/ # 中间件
│   │   └── routes/     # 路由配置
│   ├── core/           # 核心业务逻辑
│   │   ├── audio/      # 音频下载转换
│   │   ├── bilibili/   # B站API交互
│   │   ├── cache/      # 缓存管理
│   │   └── config/     # 配置管理
│   ├── models/         # 数据模型
│   └── utils/          # 工具函数
├── configs/            # 配置文件
├── parse_cache/        # 音频缓存目录
└── logs/              # 日志目录
```

### 核心模块说明

- **WBI签名模块**: 实现B站WBI签名算法，保证API请求的合法性
- **音频下载器**: 下载m4s格式音频并使用FFmpeg转换为MP3
- **缓存管理器**: 管理本地文件缓存和数据库记录
- **静态文件服务**: 提供MP3文件的HTTP访问服务

### 贡献代码

1. Fork 项目
2. 创建特性分支
3. 提交代码
4. 发起 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 支持

如果这个项目对你有帮助，欢迎 ⭐ Star 支持！

---

**免责声明**: 本项目仅供学习交流使用，请勿用于商业用途。使用者需自行承担使用风险。
