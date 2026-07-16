# OIDC Proxy Dockerfile
# 多阶段构建喵

# ── 编译阶段 ──
FROM golang:1.21-alpine AS builder

# 安装 build 依赖喵
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存喵
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译喵
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o oidc-proxy .

# ── 运行阶段 ──
FROM alpine:3.19

# 创建非 root 用户喵
RUN adduser -D -h /app appuser

WORKDIR /app

# 复制二进制文件喵
COPY --from=builder /app/oidc-proxy .
COPY --from=builder /app/config.yaml .

# 创建数据目录和密钥目录喵
RUN mkdir -p keys data && chown -R appuser:appuser /app

USER appuser

# 暴露端口喵
EXPOSE 8443

# 启动命令喵
CMD ["./oidc-proxy", "-config", "config.yaml"]
