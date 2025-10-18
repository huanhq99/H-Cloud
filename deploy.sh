#!/bin/bash

# H-Cloud 云盘部署脚本
# 适用于 VPS 生产环境部署

set -e

echo "🚀 开始部署 H-Cloud 云盘..."

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    echo "安装命令: curl -fsSL https://get.docker.com | sh"
    exit 1
fi

# 检查 Docker Compose 是否安装
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

# 创建必要的目录
echo "📁 创建数据目录..."
mkdir -p data storage logs

# 设置目录权限
chmod 755 data storage logs

# 检查 .env 文件
if [ ! -f .env ]; then
    echo "⚠️  .env 文件不存在，从模板创建..."
    if [ -f .env.example ]; then
        cp .env.example .env
        echo "✅ 已创建 .env 文件，请编辑配置后重新运行部署脚本"
        echo "重要: 请修改以下配置项："
        echo "  - ADMIN_PASSWORD (管理员密码)"
        echo "  - JWT_SECRET (JWT 密钥，至少32位字符)"
        echo "  - PORT (如需要修改端口)"
        exit 1
    else
        echo "❌ .env.example 文件不存在"
        exit 1
    fi
fi

# 检查关键配置
echo "🔍 检查配置..."
source .env

if [ "$ADMIN_PASSWORD" = "your_secure_password_here" ]; then
    echo "❌ 请修改 .env 文件中的 ADMIN_PASSWORD"
    exit 1
fi

if [ "$JWT_SECRET" = "your_jwt_secret_key_here_at_least_32_characters" ]; then
    echo "❌ 请修改 .env 文件中的 JWT_SECRET"
    exit 1
fi

# 停止现有容器
echo "🛑 停止现有容器..."
docker-compose down 2>/dev/null || docker compose down 2>/dev/null || true

# 构建并启动服务
echo "🔨 构建镜像..."
if command -v docker-compose &> /dev/null; then
    docker-compose build --no-cache
    echo "🚀 启动服务..."
    docker-compose up -d
else
    docker compose build --no-cache
    echo "🚀 启动服务..."
    docker compose up -d
fi

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
if command -v docker-compose &> /dev/null; then
    docker-compose ps
else
    docker compose ps
fi

# 检查健康状态
echo "🔍 检查服务健康状态..."
for i in {1..30}; do
    if curl -f http://localhost:${PORT:-8080}/api/system/info >/dev/null 2>&1; then
        echo "✅ 服务启动成功！"
        echo ""
        echo "🎉 部署完成！"
        echo "📱 访问地址: http://your-server-ip:${PORT:-8080}/login.html"
        echo "👤 管理员账号: $ADMIN_USERNAME"
        echo "🔑 管理员密码: $ADMIN_PASSWORD"
        echo ""
        echo "📋 常用命令:"
        echo "  查看日志: docker-compose logs -f h-cloud"
        echo "  重启服务: docker-compose restart h-cloud"
        echo "  停止服务: docker-compose down"
        echo "  更新服务: ./deploy.sh"
        exit 0
    fi
    echo "等待服务启动... ($i/30)"
    sleep 2
done

echo "❌ 服务启动失败，请检查日志:"
if command -v docker-compose &> /dev/null; then
    docker-compose logs h-cloud
else
    docker compose logs h-cloud
fi
exit 1