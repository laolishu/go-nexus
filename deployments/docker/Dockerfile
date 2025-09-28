# 多阶段构建
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 安装Wire
RUN go install github.com/google/wire/cmd/wire@latest

# 生成Wire代码
RUN cd cmd/server && wire

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags '-X main.Version=docker -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) -s -w' \
    -o go-nexus cmd/server/main.go cmd/server/wire_gen.go

# 运行时镜像
FROM alpine:3.18

# 安装必要工具和证书
RUN apk --no-cache add ca-certificates tzdata wget && \
    rm -rf /var/cache/apk/*

# 创建用户和目录
RUN adduser -D -s /bin/sh nexus && \
    mkdir -p /var/lib/go-nexus /var/log/go-nexus && \
    chown -R nexus:nexus /var/lib/go-nexus /var/log/go-nexus

# 切换到应用用户
USER nexus

WORKDIR /app

# 复制二进制文件和配置
COPY --from=builder /app/go-nexus .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/web ./web

# 创建数据目录
RUN mkdir -p /var/lib/go-nexus/data /var/lib/go-nexus/plugins

# 暴露端口
EXPOSE 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# 设置环境变量
ENV GIN_MODE=release

# 启动命令
CMD ["./go-nexus", "server", "--config", "configs/config.yaml"]