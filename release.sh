#!/bin/bash

# H-Cloud 自动化发布脚本
# 支持版本管理、Docker构建、GitHub推送和Docker Hub发布

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
DOCKER_USERNAME="huanhq99"
DOCKER_REPO="h-cloud"
GITHUB_REPO="huanhq99/H-Cloud"

# 函数：打印彩色信息
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

# 函数：获取当前版本号
get_current_version() {
    if [ -f "VERSION" ]; then
        cat VERSION
    else
        echo "v1.2.3"
    fi
}

# 函数：更新版本号
update_version() {
    local new_version=$1
    echo "$new_version" > VERSION
    
    # 更新 docker-compose.yml 中的版本
    if [ -f "docker-compose.yml" ]; then
        sed -i.bak "s/image: ${DOCKER_USERNAME}\/${DOCKER_REPO}:.*/image: ${DOCKER_USERNAME}\/${DOCKER_REPO}:${new_version}/" docker-compose.yml
        rm docker-compose.yml.bak
        print_success "已更新 docker-compose.yml 中的版本为 $new_version"
    fi
    
    # 更新系统控制器中的版本号
    if [ -f "backend/internal/api/system_controller.go" ]; then
        sed -i.bak "s/\"version\": \".*\"/\"version\": \"${new_version}\"/" backend/internal/api/system_controller.go
        rm backend/internal/api/system_controller.go.bak
        print_success "已更新系统控制器中的版本为 $new_version"
    fi
    
    # 更新前端页面中的版本号
    if [ -f "backend/public/api.html" ]; then
        sed -i.bak "s/value=\"v[0-9]\+\.[0-9]\+\.[0-9]\+\"/value=\"${new_version}\"/" backend/public/api.html
        rm backend/public/api.html.bak
        print_success "已更新前端页面中的版本为 $new_version"
    fi
}

# 函数：构建 Docker 镜像
build_docker_image() {
    local version=$1
    local build_date=$(date +"%Y-%m-%d")
    
    print_info "开始构建 AMD64 Docker 镜像..."
    
    # 构建 AMD64 镜像
    docker build --platform linux/amd64 \
        --build-arg VERSION="$version" \
        --build-arg BUILD_DATE="$build_date" \
        -f amd64-dockerfile \
        -t "${DOCKER_USERNAME}/${DOCKER_REPO}:${version}" \
        -t "${DOCKER_USERNAME}/${DOCKER_REPO}:latest" \
        .
    
    print_success "Docker 镜像构建完成: ${DOCKER_USERNAME}/${DOCKER_REPO}:${version}"
}

# 函数：推送到 Docker Hub
push_to_docker_hub() {
    local version=$1
    
    print_info "推送镜像到 Docker Hub..."
    
    # 推送版本标签
    docker push "${DOCKER_USERNAME}/${DOCKER_REPO}:${version}"
    print_success "已推送版本标签: ${version}"
    
    # 推送 latest 标签
    docker push "${DOCKER_USERNAME}/${DOCKER_REPO}:latest"
    print_success "已推送 latest 标签"
}

# 函数：提交到 GitHub
commit_and_push_github() {
    local version=$1
    local message="Release ${version}"
    
    print_info "提交代码到 GitHub..."
    
    # 添加所有更改
    git add .
    
    # 检查是否有更改需要提交
    if git diff --staged --quiet; then
        print_warning "没有检测到代码更改，跳过提交"
        return
    fi
    
    # 提交更改
    git commit -m "$message"
    
    # 创建标签
    git tag -a "$version" -m "Release $version"
    
    # 推送到远程仓库
    git push origin main
    git push origin "$version"
    
    print_success "已推送到 GitHub: $version"
}

# 函数：验证发布
verify_release() {
    local version=$1
    
    print_info "验证发布..."
    
    # 检查 Docker Hub 上的镜像
    print_info "检查 Docker Hub 镜像..."
    if docker manifest inspect "${DOCKER_USERNAME}/${DOCKER_REPO}:${version}" > /dev/null 2>&1; then
        print_success "Docker Hub 镜像验证成功"
    else
        print_error "Docker Hub 镜像验证失败"
        return 1
    fi
    
    # 测试拉取镜像
    print_info "测试拉取最新镜像..."
    docker pull "${DOCKER_USERNAME}/${DOCKER_REPO}:${version}"
    print_success "镜像拉取测试成功"
}

# 函数：显示帮助信息
show_help() {
    echo "H-Cloud 自动化发布脚本"
    echo ""
    echo "用法: $0 [选项] <版本号>"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示帮助信息"
    echo "  -s, --skip-git 跳过 Git 操作"
    echo "  -d, --dry-run  预演模式，不执行实际操作"
    echo ""
    echo "示例:"
    echo "  $0 v1.2.4              # 发布版本 v1.2.4"
    echo "  $0 --skip-git v1.2.4   # 发布版本但跳过 Git 操作"
    echo "  $0 --dry-run v1.2.4    # 预演发布流程"
}

# 主函数
main() {
    local skip_git=false
    local dry_run=false
    local version=""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -s|--skip-git)
                skip_git=true
                shift
                ;;
            -d|--dry-run)
                dry_run=true
                shift
                ;;
            v*)
                version=$1
                shift
                ;;
            *)
                print_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 检查版本号
    if [ -z "$version" ]; then
        current_version=$(get_current_version)
        print_info "当前版本: $current_version"
        read -p "请输入新版本号 (例如: v1.2.4): " version
        
        if [ -z "$version" ]; then
            print_error "版本号不能为空"
            exit 1
        fi
    fi
    
    # 验证版本号格式
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        print_error "版本号格式错误，应为 vX.Y.Z 格式"
        exit 1
    fi
    
    print_info "开始发布流程: $version"
    
    if [ "$dry_run" = true ]; then
        print_warning "预演模式 - 不会执行实际操作"
        print_info "将要执行的操作:"
        print_info "1. 更新版本号到 $version"
        print_info "2. 构建 Docker 镜像"
        print_info "3. 推送到 Docker Hub"
        if [ "$skip_git" = false ]; then
            print_info "4. 提交并推送到 GitHub"
        fi
        print_info "5. 验证发布"
        exit 0
    fi
    
    # 检查必要工具
    for tool in docker git; do
        if ! command -v $tool &> /dev/null; then
            print_error "$tool 未安装或不在 PATH 中"
            exit 1
        fi
    done
    
    # 检查 Docker 登录状态
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker 未运行或无权限访问"
        exit 1
    fi
    
    # 检查 Docker Hub 登录状态
    print_info "检查 Docker Hub 登录状态..."
    if ! docker system info 2>/dev/null | grep -q "Username:" && ! cat ~/.docker/config.json 2>/dev/null | grep -q "https://index.docker.io/v1/"; then
        print_warning "需要登录 Docker Hub"
        print_info "请运行以下命令登录:"
        echo "echo 'YOUR_PASSWORD_OR_TOKEN' | docker login --username ${DOCKER_USERNAME} --password-stdin"
        print_info "或者使用交互式登录:"
        echo "docker login --username ${DOCKER_USERNAME}"
        print_info "登录完成后重新运行此脚本"
        exit 1
    fi
    print_success "Docker Hub 登录状态正常"
    
    # 执行发布流程
    update_version "$version"
    build_docker_image "$version"
    push_to_docker_hub "$version"
    
    if [ "$skip_git" = false ]; then
        commit_and_push_github "$version"
    fi
    
    verify_release "$version"
    
    print_success "🎉 发布完成！"
    print_info "版本: $version"
    print_info "Docker Hub: https://hub.docker.com/r/${DOCKER_USERNAME}/${DOCKER_REPO}"
    if [ "$skip_git" = false ]; then
        print_info "GitHub: https://github.com/${GITHUB_REPO}/releases/tag/${version}"
    fi
    
    print_info "使用以下命令拉取最新版本:"
    echo "docker pull ${DOCKER_USERNAME}/${DOCKER_REPO}:${version}"
    echo "docker pull ${DOCKER_USERNAME}/${DOCKER_REPO}:latest"
}

# 执行主函数
main "$@"