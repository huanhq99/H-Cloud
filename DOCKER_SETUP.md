# Docker 设置和发布指南

## 问题诊断

如果遇到 Docker 推送失败的问题，通常是由于以下原因：

1. **未登录 Docker Hub**
2. **登录凭据过期**
3. **网络连接问题**
4. **权限问题**

## 解决方案

### 1. 使用自动登录脚本

我们提供了一个自动化的 Docker 登录脚本：

```bash
./docker-login.sh
```

这个脚本会：
- 检查 Docker 运行状态
- 验证当前登录状态
- 引导你完成登录过程
- 验证登录是否成功

### 2. 手动登录

如果自动脚本不工作，可以手动登录：

```bash
# 交互式登录
docker login --username huanhq99

# 或使用 Personal Access Token（推荐）
echo 'YOUR_TOKEN' | docker login --username huanhq99 --password-stdin
```

### 3. 创建 Personal Access Token

为了更安全的登录，建议使用 Personal Access Token：

1. 访问 [Docker Hub Settings](https://hub.docker.com/settings/security)
2. 点击 "New Access Token"
3. 输入描述（如 "H-Cloud Release"）
4. 选择权限（Read, Write, Delete）
5. 复制生成的 token
6. 使用 token 登录：
   ```bash
   echo 'YOUR_TOKEN' | docker login --username huanhq99 --password-stdin
   ```

## 发布流程

登录成功后，可以使用发布脚本：

```bash
# 发布新版本
./release.sh v1.2.4

# 预览发布流程（不执行实际操作）
./release.sh --dry-run v1.2.4

# 跳过 Git 操作，只发布 Docker 镜像
./release.sh --skip-git v1.2.4
```

## 验证发布

发布完成后，可以验证：

```bash
# 拉取最新版本
docker pull huanhq99/h-cloud:v1.2.4
docker pull huanhq99/h-cloud:latest

# 检查镜像信息
docker inspect huanhq99/h-cloud:v1.2.4
```

## 常见问题

### Q: 登录时提示 "unauthorized: incorrect username or password"
A: 
- 确认用户名是 `huanhq99`
- 如果使用密码，确保密码正确
- 推荐使用 Personal Access Token 替代密码

### Q: 推送时提示 "denied: requested access to the resource is denied"
A:
- 确保已正确登录到 Docker Hub
- 确认有推送权限到 `huanhq99/h-cloud` 仓库

### Q: 网络连接问题
A:
- 检查网络连接
- 如果在中国大陆，可能需要配置 Docker 镜像加速器
- 尝试使用 VPN 或其他网络环境

## 技术细节

### Docker 镜像构建

我们的构建过程包括：
- 多阶段构建优化镜像大小
- AMD64 架构支持
- 动态版本号和构建时间
- 环境变量注入

### 版本管理

- 版本号格式：`vX.Y.Z`（如 `v1.2.3`）
- 自动更新 `docker-compose.yml`
- 同步推送到 GitHub 和 Docker Hub
- 创建 Git 标签和 GitHub Release

### 自动化 CI/CD

GitHub Actions 工作流 (`.github/workflows/release.yml`) 提供：
- 自动构建和推送
- 版本号验证
- 多平台支持准备
- 发布验证

## 支持

如果仍然遇到问题，请检查：
1. Docker Desktop 是否正常运行
2. 网络连接是否稳定
3. Docker Hub 账户状态
4. 仓库权限设置