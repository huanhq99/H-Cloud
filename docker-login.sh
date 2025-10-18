#!/bin/bash

# Docker Hub 登录脚本
# 使用方法: ./docker-login.sh

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
DOCKER_USERNAME="huanhq99"

# 函数：彩色信息打印
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 检查Docker是否运行
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker 未运行或无权限访问"
        print_info "请启动 Docker Desktop 或检查 Docker 服务状态"
        exit 1
    fi
    print_success "Docker 运行正常"
}

# 检查当前登录状态
check_login_status() {
    print_info "检查当前 Docker Hub 登录状态..."
    
    # 尝试拉取一个小的公共镜像来测试登录状态
    if docker pull alpine:latest > /dev/null 2>&1; then
        print_success "Docker Hub 连接正常"
        
        # 检查是否已登录到指定用户
        if docker info 2>/dev/null | grep -q "Username: ${DOCKER_USERNAME}"; then
            print_success "已登录到 Docker Hub (用户: ${DOCKER_USERNAME})"
            return 0
        else
            print_warning "未登录到 Docker Hub 或登录用户不匹配"
            return 1
        fi
    else
        print_error "无法连接到 Docker Hub"
        return 1
    fi
}

# 执行登录
perform_login() {
    print_info "开始登录到 Docker Hub..."
    print_info "用户名: ${DOCKER_USERNAME}"
    print_warning "请输入密码或 Personal Access Token:"
    
    if docker login --username "${DOCKER_USERNAME}"; then
        print_success "登录成功！"
        return 0
    else
        print_error "登录失败"
        return 1
    fi
}

# 验证登录
verify_login() {
    print_info "验证登录状态..."
    
    # 尝试推送一个测试标签（不会实际推送）
    if docker tag alpine:latest "${DOCKER_USERNAME}/test:login-test" 2>/dev/null; then
        docker rmi "${DOCKER_USERNAME}/test:login-test" 2>/dev/null
        print_success "登录验证成功"
        return 0
    else
        print_error "登录验证失败"
        return 1
    fi
}

# 主函数
main() {
    echo "=== Docker Hub 登录工具 ==="
    echo ""
    
    # 检查Docker
    check_docker
    
    # 检查登录状态
    if check_login_status; then
        print_info "已经登录，无需重新登录"
        exit 0
    fi
    
    # 执行登录
    if perform_login; then
        verify_login
        print_success "🎉 Docker Hub 登录完成！"
        print_info "现在可以运行发布脚本了:"
        echo "./release.sh v1.2.4"
    else
        print_error "登录失败，请检查用户名和密码"
        print_info "提示："
        print_info "1. 确保用户名正确: ${DOCKER_USERNAME}"
        print_info "2. 使用 Docker Hub 密码或 Personal Access Token"
        print_info "3. 创建 Personal Access Token: https://hub.docker.com/settings/security"
        exit 1
    fi
}

# 执行主函数
main "$@"