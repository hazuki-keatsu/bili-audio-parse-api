FROM golang:1.21-alpine AS builder

# 安装必要的工具
RUN apk --no-cache add ca-certificates git

WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata sqlite

WORKDIR /root/

# 从构建阶段复制可执行文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# 创建必要的目录
RUN mkdir -p parse_cache logs

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"]
