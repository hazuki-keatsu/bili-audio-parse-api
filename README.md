# B站DASH音频解析服务

一个基于Go语言开发的B站音频解析后端服务，使用Gin Web框架和GORM ORM框架，采用SQLite作为数据库。

> :heavy_exclamation_mark: __注意__
> 
> 本项目所有的代码都由AI编写，本项目是作者对AI编程工具流的尝试。
>
> - [需求文档](./RequirementDocuments.md) 由 豆包 撰写
> - [开发文档](./DevelopmentDocuments.md) 由 Claude Sonnet 4 撰写
> - 项目所有代码由 Claude Sonnet 4 撰写

## 项目特性

- 🎵 **音频解析**：支持通过BV号获取DASH格式音频资源
- 🔐 **多登录方式**：支持账号密码登录和SESSDATA登录
- 🚀 **高性能**：内置缓存机制，减少重复请求
- 🛡️ **安全可靠**：JWT认证、敏感信息加密存储
- 📊 **监控统计**：提供服务状态和统计信息接口
- 🔄 **限流保护**：防止API滥用

## 技术栈

- **编程语言**：Go 1.25.1
- **Web框架**：Gin v1.10.1
- **ORM框架**：GORM v2.0+
- **数据库**：SQLite 3
- **认证**：JWT
- **加密**：AES

## 快速开始

### 环境要求

- Go 1.19+
- Git

### 安装依赖

```bash
go mod tidy
```

### 配置环境变量

复制 `.env` 文件并根据需要修改配置：

```bash
cp .env .env.local
```

### 运行服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

## API接口

### 1. 账号密码登录

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"your_username","password":"your_password"}'
```

### 2. SESSDATA登录

```bash
curl -X POST http://localhost:8080/api/login/sessdata \
  -H "Content-Type: application/json" \
  -d '{"sessdata":"your_sessdata","bili_jct":"your_bili_jct","dedeuserid":"your_dedeuserid"}'
```

### 3. 解析音频

```bash
curl -X GET "http://localhost:8080/api/parse?bv=BV1234567890&quality=0" \
  -H "Authorization: Bearer your_jwt_token"
```

### 4. 服务状态

```bash
curl -X GET http://localhost:8080/api/status
```

### 5. 健康检查

```bash
curl -X GET http://localhost:8080/health
```

## 项目结构

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
│   └── bilibili.go
├── middlewares/            # 中间件
│   ├── auth.go
│   └── ratelimit.go
├── utils/                  # 工具函数
│   ├── crypto.go
│   └── wbi.go
├── routes/                 # 路由配置
│   └── routes.go
├── logs/                   # 日志文件
├── data/                   # 数据库文件
└── README.md
```

## 部署

### 编译二进制文件

```bash
# 本地编译
go build -o bili-parse-api main.go

# Windows交叉编译
GOOS=windows GOARCH=amd64 go build -o bili-parse-api.exe main.go

# Linux交叉编译
GOOS=linux GOARCH=amd64 go build -o bili-parse-api main.go
```

### Docker部署

```bash
# 构建镜像
docker build -t bili-parse-api .

# 运行容器
docker run -d -p 8080:8080 \
  -e JWT_SECRET=your-secret-key \
  -e ENCRYPTION_KEY=your-32-byte-key \
  -v ./data:/app/data \
  bili-parse-api
```

## 配置说明

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| SERVER_HOST | 0.0.0.0 | 服务器监听地址 |
| SERVER_PORT | 8080 | 服务器端口 |
| GIN_MODE | debug | Gin运行模式 |
| DB_PATH | ./data/bili.db | 数据库文件路径 |
| JWT_SECRET | your-secret-key | JWT签名密钥 |
| ENCRYPTION_KEY | your-32-byte-key-here-123456789 | AES加密密钥(32字节) |
| CACHE_TTL | 3600 | 缓存时间(秒) |

## 注意事项

1. **安全性**：
   - 请务必修改默认的JWT密钥和加密密钥
   - 生产环境建议使用HTTPS
   - 定期更新密钥

2. **合规性**：
   - 请遵守B站服务条款
   - 尊重版权保护
   - 避免过度请求

3. **性能**：
   - 合理设置缓存时间
   - 监控API请求频率
   - 定期清理过期数据

## 开发指南

### 本地开发

1. 克隆项目
2. 安装依赖：`go mod tidy`
3. 配置环境变量
4. 运行：`go run main.go`

### 代码规范

- 使用Go官方代码格式化工具：`go fmt`
- 运行静态检查：`go vet`
- 遵循Go代码规范

## 许可证

本项目仅用于学习和研究目的，请遵守相关法律法规和B站服务条款。

## 支持

如有问题，请提交Issue或Pull Request。
