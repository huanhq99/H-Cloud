FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制go mod和sum文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -o h-cloud-server ./cmd/server

# 最终镜像
FROM alpine:3.18

WORKDIR /app

# 安装必要的系统工具
RUN apk --no-cache add ca-certificates tzdata curl

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