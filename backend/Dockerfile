# 强制使用AMD64架构 - 绝对不要ARM64！
# syntax=docker/dockerfile:1
FROM --platform=linux/amd64 golang:1.24-alpine AS builder

WORKDIR /app

# 安装编译依赖（SQLite需要CGO和gcc）
RUN apk --no-cache add gcc musl-dev sqlite-dev

# 复制go mod和sum文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译（启用CGO以支持SQLite，强制指定AMD64架构）
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o h-cloud-server ./cmd/server

# 最终镜像 - 强制使用AMD64架构
FROM --platform=linux/amd64 alpine:3.18

WORKDIR /app

# 安装运行时依赖（SQLite运行时库）
RUN apk --no-cache add ca-certificates tzdata curl sqlite

# 从builder阶段复制编译好的应用
COPY --from=builder /app/h-cloud-server .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/public ./public

# 创建存储目录
RUN mkdir -p /data/storage /app/storage

# 设置时区
ENV TZ=Asia/Shanghai

EXPOSE 8080

CMD ["./h-cloud-server"]