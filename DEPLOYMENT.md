# H-Cloud Drive 部署指南

本文档提供了 H-Cloud Drive 的详细部署说明。

## 目录
- [系统要求](#系统要求)
- [VPS 一键部署](#vps-一键部署)
- [Docker 部署](#docker-部署)
- [手动部署](#手动部署)
- [配置说明](#配置说明)
- [故障排除](#故障排除)

## 系统要求

### 最低配置
- CPU: 1 核心
- 内存: 512MB
- 存储: 10GB 可用空间
- 操作系统: Linux/macOS/Windows

### 推荐配置
- CPU: 2 核心或以上
- 内存: 2GB 或以上
- 存储: 50GB 或以上可用空间
- 操作系统: Ubuntu 20.04+ / CentOS 8+ / macOS 10.15+

## VPS 一键部署

### 前置要求
- Docker 20.10+
- Docker Compose 2.0+

### 快速部署

1. 克隆项目
```bash
git clone https://github.com/huanhq99/H-Cloud.git
cd H-Cloud
```

2. 运行一键部署脚本
```bash
./deploy.sh
```

脚本会自动：
- 检查 Docker 环境
- 创建必要的目录
- 生成配置文件模板
- 构建并启动服务
- 检查服务健康状态

3. 首次运行时，脚本会提示你编辑 `.env` 文件：
```bash
# 编辑配置文件
nano .env

# 重要：必须修改以下配置
ADMIN_PASSWORD=your_secure_password_here
JWT_SECRET=your_jwt_secret_key_here_at_least_32_characters
PORT=8080  # 可选，默认8080
```

4. 再次运行部署脚本
```bash
./deploy.sh
```

5. 访问服务
打开浏览器访问 `http://your-server-ip:8080/login.html`

### 生产环境配置

#### 安全配置
```bash
# .env 文件示例
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your_very_secure_password_123!
JWT_SECRET=your_super_secret_jwt_key_at_least_32_characters_long
GIN_MODE=release
LOG_LEVEL=info
PORT=8080
```

#### 防火墙配置
```bash
# Ubuntu/Debian
sudo ufw allow 8080/tcp

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

#### 反向代理 (Nginx)
```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Docker 部署

### 前置要求
- Docker 20.10+
- Docker Compose 2.0+ (可选)

### 方式一：使用 Docker Hub 镜像 (推荐)

#### 快速部署
```bash
# 1. 创建数据目录
mkdir -p h-cloud-data/data h-cloud-data/storage

# 2. 直接运行容器
docker run -d \
  --name h-cloud \
  -p 8080:8080 \
  -v $(pwd)/h-cloud-data/data:/data \
  -v $(pwd)/h-cloud-data/storage:/app/storage \
  -e ADMIN_USERNAME=admin \
  -e ADMIN_PASSWORD=your_secure_password \
  -e JWT_SECRET=your_jwt_secret_key_at_least_32_characters \
  -e GIN_MODE=release \
  --restart unless-stopped \
  huanhq99/h-cloud:latest

# 3. 查看容器状态
docker ps

# 4. 查看日志
docker logs -f h-cloud
```

#### 生产环境配置
```bash
# 创建配置文件
cat > h-cloud.env << EOF
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your_very_secure_password_123!
JWT_SECRET=your_super_secret_jwt_key_at_least_32_characters_long
GIN_MODE=release
LOG_LEVEL=info
PORT=8080
EOF

# 使用配置文件运行
docker run -d \
  --name h-cloud \
  -p 8080:8080 \
  -v $(pwd)/h-cloud-data/data:/data \
  -v $(pwd)/h-cloud-data/storage:/app/storage \
  --env-file h-cloud.env \
  --restart unless-stopped \
  huanhq99/h-cloud:latest
```

### 方式二：使用 Docker Compose

#### 手动部署

1. 克隆项目
```bash
git clone https://github.com/huanhq99/H-Cloud.git
cd H-Cloud
```

2. 配置环境变量
```bash
cp .env.example .env
# 编辑 .env 文件，修改必要的配置
```

3. 启动服务
```bash
docker-compose up -d
```

4. 访问服务
打开浏览器访问 `http://localhost:8080/login.html`

### 详细配置

#### 环境变量说明
- `ADMIN_USERNAME`: 管理员用户名（默认: admin）
- `ADMIN_PASSWORD`: 管理员密码（默认: admin123）
- `JWT_SECRET`: JWT 密钥（生产环境必须修改）
- `GIN_MODE`: 运行模式（debug/release）
- `LOG_LEVEL`: 日志级别（debug/info/warn/error）
- `MAX_UPLOAD_SIZE`: 最大上传文件大小
- `PORT`: 服务端口

#### 数据持久化
默认情况下，数据存储在以下目录：
- `./data`: 应用数据
- `./storage`: 用户文件存储
- `./logs`: 应用日志

## 手动部署

### 前置要求
- Go 1.21+
- Node.js 16+ (如果需要构建前端)

### 后端部署

1. 编译后端
```bash
cd backend
go mod download
go build -o h-cloud-server cmd/server/main.go
```

2. 配置文件
```bash
cp configs/config.yaml.example configs/config.yaml
# 编辑配置文件
```

3. 启动服务
```bash
./h-cloud-server
```

## 配置说明

### 基本配置
- 服务端口: 8080
- 数据目录: ./data
- 存储目录: ./storage
- 日志目录: ./logs

### 安全配置
- 修改默认管理员密码
- 设置强 JWT 密钥（至少32位字符）
- 配置 HTTPS（生产环境推荐）
- 启用防火墙规则

### 性能配置
- 内存限制: 512MB（可调整）
- CPU 限制: 0.5 核心（可调整）
- 日志轮转: 10MB/文件，保留3个文件

## 故障排除

### 常见问题

1. **端口被占用**
   - 修改 .env 文件中的 PORT 配置
   - 或停止占用端口的服务

2. **权限问题**
   - 确保数据目录有正确的读写权限
   - Linux 下可能需要调整 SELinux 设置

3. **内存不足**
   - 增加服务器内存
   - 或调整 docker-compose.yml 中的内存限制

4. **服务无法启动**
   - 检查 .env 配置是否正确
   - 查看容器日志排查问题

### 日志查看
```bash
# Docker 部署
docker-compose logs -f h-cloud

# 查看实时日志
tail -f logs/app.log
```

### 常用管理命令
```bash
# 查看服务状态
docker-compose ps

# 重启服务
docker-compose restart h-cloud

# 停止服务
docker-compose down

# 更新服务
./deploy.sh
```

## 更新升级

### 使用部署脚本更新
```bash
git pull
./deploy.sh
```

### 手动更新
```bash
git pull
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## 备份恢复

### 数据备份
```bash
# 备份用户数据
tar -czf backup-$(date +%Y%m%d).tar.gz data/ storage/ logs/
```

### 数据恢复
```bash
# 恢复数据
tar -xzf backup-YYYYMMDD.tar.gz
```

## 监控和维护

### 健康检查
服务提供健康检查接口：`http://localhost:8080/api/system/info`

### 日志监控
- 应用日志：`./logs/app.log`
- 容器日志：`docker-compose logs h-cloud`

### 定期维护
- 定期备份数据
- 监控磁盘空间使用
- 检查日志文件大小
- 更新系统和 Docker 镜像

## 🔧 部署方式

### 方式一：源码部署

#### 1. 环境准备

```bash
# 安装 Go (以 Ubuntu 为例)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

#### 2. 获取源码

```bash
# 克隆项目
git clone https://github.com/huanhq99/H-Cloud.git
cd HQyun/backend

# 安装依赖
go mod tidy
```

#### 3. 配置文件

创建或编辑 `configs/config.yaml`:

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 60
  write_timeout: 60

database:
  type: sqlite
  host: sqlite
  port: 0
  name: hqyun.db
  user: ""
  password: ""

storage:
  path: "./local_storage"
  max_file_size: 104857600  # 100MB
  allowed_types: ["jpg", "jpeg", "png", "gif", "pdf", "txt", "doc", "docx"]

admin:
  username: "admin"
  password: "your_secure_password"  # 请修改默认密码

jwt:
  secret: "your_jwt_secret_key"  # 请使用强密码
  expires_hours: 24

logging:
  level: "info"
  file: "./logs/app.log"
```

#### 4. 启动服务

```bash
# 开发模式
go run cmd/server/main.go

# 生产模式 - 编译后运行
go build -o hqyun cmd/server/main.go
./hqyun
```

#### 5. 系统服务配置 (可选)

创建 systemd 服务文件 `/etc/systemd/system/hqyun.service`:

```ini
[Unit]
Description=H-Cloud Drive
After=network.target

[Service]
Type=simple
User=hqyun
WorkingDirectory=/opt/hqyun
ExecStart=/opt/hqyun/hqyun
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启用服务:

```bash
sudo systemctl daemon-reload
sudo systemctl enable hqyun
sudo systemctl start hqyun
```

### 方式二：Docker 部署

#### 1. 使用 Docker Compose (推荐)

##### 基础部署

```bash
# 克隆项目
git clone https://github.com/huanhq99/H-yunpan.git
cd H-yunpan

# 启动服务 (仅后端)
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f hyun-backend
```

##### 带 Nginx 反向代理部署

```bash
# 启动服务 (包含 Nginx)
docker-compose --profile nginx up -d

# 查看所有服务状态
docker-compose --profile nginx ps
```

##### 环境变量配置

创建 `.env` 文件进行自定义配置:

```bash
# .env 文件示例
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your_secure_password_here
JWT_SECRET=your_super_secret_jwt_key_change_in_production
SERVER_MODE=release
LOCAL_STORAGE_PATH=./data/storage
```

##### 常用 Docker Compose 命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f

# 更新服务 (重新构建)
docker-compose up -d --build

# 清理数据 (谨慎使用)
docker-compose down -v
```

#### 2. 单独使用 Docker

##### 构建和运行

```bash
# 构建镜像
cd H-yunpan/backend
docker build -t h-yun-cloud-drive:latest .

# 创建数据目录
mkdir -p ./data/storage ./data/uploads

# 运行容器
docker run -d \
  --name h-yun-backend \
  -p 8080:8080 \
  -v $(pwd)/data:/data \
  -e ADMIN_USERNAME=admin \
  -e ADMIN_PASSWORD=your_secure_password \
  -e JWT_SECRET=your_jwt_secret_key \
  -e DB_PATH=/data/hyun_disk.db \
  -e STORAGE_PATH=/data/storage \
  -e UPLOAD_PATH=/data/uploads \
  --restart unless-stopped \
  h-yun-cloud-drive:latest
```

##### Docker 网络配置 (多容器)

```bash
# 创建自定义网络
docker network create hyun-network

# 运行后端服务
docker run -d \
  --name h-yun-backend \
  --network hyun-network \
  -p 8080:8080 \
  -v hyun-data:/data \
  -e ADMIN_USERNAME=admin \
  -e ADMIN_PASSWORD=your_password \
  h-yun-cloud-drive:latest

# 运行 Nginx 代理 (可选)
docker run -d \
  --name h-yun-nginx \
  --network hyun-network \
  -p 80:80 \
  -p 443:443 \
  -v $(pwd)/frontend/nginx.conf:/etc/nginx/nginx.conf:ro \
  nginx:alpine
```

### 方式三：云服务器部署

#### 腾讯云/阿里云部署示例

```bash
# 1. 购买云服务器 (1核2G即可)
# 2. 安装 Docker
curl -fsSL https://get.docker.com | bash -s docker
sudo systemctl start docker
sudo systemctl enable docker

# 3. 部署应用
git clone https://github.com/huanhq99/H-Cloud.git
cd HQyun
docker-compose up -d

# 4. 配置防火墙
sudo ufw allow 8080
sudo ufw enable
```

## ⚙️ 配置详解

### 环境变量配置

| 变量名 | 描述 | 默认值 | 示例 |
|--------|------|--------|------|
| `SERVER_PORT` | 服务端口 | 8080 | 8080 |
| `SERVER_HOST` | 服务地址 | 0.0.0.0 | 0.0.0.0 |
| `DB_PATH` | 数据库路径 | ./hqyun.db | /data/hqyun.db |
| `STORAGE_PATH` | 存储路径 | ./local_storage | /data/storage |
| `ADMIN_USERNAME` | 管理员用户名 | admin | admin |
| `ADMIN_PASSWORD` | 管理员密码 | password | your_password |
| `JWT_SECRET` | JWT 密钥 | - | your_jwt_secret |
| `MAX_FILE_SIZE` | 最大文件大小 | 104857600 | 209715200 |

### 配置文件详解

#### 服务器配置

```yaml
server:
  port: 8080                # 监听端口
  host: "0.0.0.0"          # 监听地址
  read_timeout: 60         # 读取超时 (秒)
  write_timeout: 60        # 写入超时 (秒)
  max_header_bytes: 1048576 # 最大请求头大小
```

#### 数据库配置

```yaml
database:
  type: sqlite             # 数据库类型
  name: hqyun.db          # 数据库文件名
  # 未来支持 MySQL/PostgreSQL
  # host: localhost
  # port: 3306
  # user: username
  # password: password
```

#### 存储配置

```yaml
storage:
  path: "./local_storage"  # 存储根目录
  max_file_size: 104857600 # 最大文件大小 (字节)
  allowed_types:           # 允许的文件类型
    - "jpg"
    - "jpeg"
    - "png"
    - "gif"
    - "pdf"
    - "txt"
    - "doc"
    - "docx"
    - "zip"
    - "rar"
```

#### 安全配置

```yaml
admin:
  username: "admin"        # 管理员用户名
  password: "password"     # 管理员密码 (请修改)

jwt:
  secret: "your_secret"    # JWT 签名密钥 (必须修改)
  expires_hours: 24        # Token 过期时间 (小时)

security:
  cors_origins:            # 允许的跨域源
    - "http://localhost:3000"
    - "https://yourdomain.com"
  rate_limit: 100          # 每分钟请求限制
```

## 🔒 安全配置

### 1. 修改默认密码

```bash
# 方式一：环境变量
export ADMIN_PASSWORD="your_secure_password"

# 方式二：配置文件
# 编辑 configs/config.yaml
admin:
  password: "your_secure_password"
```

### 2. 配置 HTTPS

#### 使用 Nginx 反向代理

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 3. 防火墙配置

```bash
# Ubuntu/Debian
sudo ufw allow 22          # SSH
sudo ufw allow 80          # HTTP
sudo ufw allow 443         # HTTPS
sudo ufw deny 8080         # 禁止直接访问应用端口
sudo ufw enable

# CentOS/RHEL
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

## 📊 性能优化

### 1. 数据库优化

```yaml
database:
  # SQLite 优化
  pragma:
    journal_mode: WAL      # 写前日志模式
    synchronous: NORMAL    # 同步模式
    cache_size: -64000     # 缓存大小 (KB)
    temp_store: MEMORY     # 临时存储在内存
```

### 2. 文件存储优化

```bash
# 使用 SSD 存储
# 定期清理临时文件
find ./local_storage -name "*.tmp" -mtime +1 -delete

# 文件压缩 (可选)
# 对于大文件可以启用压缩存储
```

### 3. 内存优化

```yaml
server:
  max_multipart_memory: 33554432  # 32MB 多部分表单内存限制
```

## 🔍 监控和日志

### 1. 日志配置

```yaml
logging:
  level: "info"              # 日志级别: debug, info, warn, error
  file: "./logs/app.log"     # 日志文件路径
  max_size: 100              # 最大文件大小 (MB)
  max_backups: 5             # 保留备份数量
  max_age: 30                # 保留天数
  compress: true             # 是否压缩旧日志
```

### 2. 系统监控

```bash
# 查看服务状态
systemctl status hqyun

# 查看资源使用
htop
df -h
du -sh ./local_storage

# 查看网络连接
netstat -tlnp | grep 8080
```

## 🚨 故障排除

### 常见问题

#### 1. 端口被占用

```bash
# 查看端口占用
lsof -i :8080
netstat -tlnp | grep 8080

# 杀死进程
kill -9 <PID>
```

#### 2. 权限问题

```bash
# 检查文件权限
ls -la ./local_storage
ls -la ./hqyun.db

# 修复权限
chmod 755 ./local_storage
chmod 644 ./hqyun.db
```

#### 3. 数据库锁定

```bash
# 检查数据库文件
file ./hqyun.db

# 重启服务
systemctl restart hqyun
```

#### 4. 内存不足

```bash
# 检查内存使用
free -h

# 清理缓存
sync && echo 3 > /proc/sys/vm/drop_caches
```

### 日志分析

```bash
# 查看错误日志
tail -f ./logs/app.log | grep ERROR

# 查看访问日志
tail -f ./logs/app.log | grep "GET\|POST"

# 统计错误数量
grep ERROR ./logs/app.log | wc -l
```

## 📈 扩展部署

### 负载均衡部署

```yaml
# docker-compose.yml
version: '3.8'
services:
  hqyun-1:
    build: ./backend
    ports:
      - "8081:8080"
    volumes:
      - ./shared_storage:/app/local_storage

  hqyun-2:
    build: ./backend
    ports:
      - "8082:8080"
    volumes:
      - ./shared_storage:/app/local_storage

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
```

### 数据备份

```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/hqyun_$DATE"

mkdir -p $BACKUP_DIR
cp ./hqyun.db $BACKUP_DIR/
cp -r ./local_storage $BACKUP_DIR/
cp -r ./configs $BACKUP_DIR/

tar -czf "/backup/hqyun_backup_$DATE.tar.gz" -C /backup "hqyun_$DATE"
rm -rf $BACKUP_DIR

# 保留最近 7 天的备份
find /backup -name "hqyun_backup_*.tar.gz" -mtime +7 -delete
```

---

📞 **需要帮助？** 请查看 [Issues](https://github.com/huanhq99/H-Cloud/issues) 或提交新的问题报告。