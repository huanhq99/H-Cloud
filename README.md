# H-Yun Cloud Drive 🌤️

一个基于 Go + SQLite 的轻量级云盘系统，支持文件上传、下载、分享和管理功能。

## ✨ 功能特性

### 🔐 用户系统
- 用户注册、登录、JWT 认证
- 管理员登录系统
- 安全的密码加密存储

### 📁 文件管理
- 文件上传、下载、删除
- 文件夹创建、管理
- 文件预览（图片、文本等）
- 批量操作支持

### 🔗 分享功能
- 文件/文件夹分享链接生成
- 分享权限控制（公开/私密）
- 分享链接过期时间设置
- 访问密码保护

### 🖼️ 图床功能
- 图片直链生成
- 多种图片格式支持
- 图片压缩和优化

### 🎨 用户界面
- 现代化响应式设计
- 深色/浅色主题切换
- 拖拽上传支持
- 移动端友好

### ⚙️ 系统管理
- 管理员后台
- 系统信息监控
- 用户管理
- 存储空间统计

## 🚀 快速开始

### 环境要求

- Go 1.19+
- SQLite 3
- 现代浏览器

### 安装部署

#### 1. 克隆项目

```bash
git clone https://github.com/yourusername/HQyun.git
cd HQyun
```

#### 2. 后端部署

```bash
cd backend

# 安装依赖
go mod tidy

# 运行服务
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动

#### 3. 使用 Docker 部署

```bash
# 构建并启动服务
docker-compose up -d
```

### 默认管理员账号

- 用户名: `admin`
- 密码: `password`

> ⚠️ **安全提醒**: 首次部署后请立即修改默认密码！

## 📖 配置说明

### 环境变量配置

```bash
# 管理员账号配置
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your_secure_password

# 数据库配置
DB_PATH=./hqyun.db

# 服务器配置
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# JWT 配置
JWT_SECRET=your_jwt_secret_key
```

### 配置文件

编辑 `backend/configs/config.yaml`:

```yaml
server:
  port: 8080
  host: "0.0.0.0"

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

admin:
  username: "admin"
  password: "password"
```

## 🔧 API 接口

### 认证接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/auth/register` | 用户注册 |
| POST | `/api/auth/login` | 用户登录 |
| GET | `/api/auth/me` | 获取当前用户信息 |

### 管理员接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/admin/login` | 管理员登录 |
| POST | `/api/admin/logout` | 管理员登出 |
| GET | `/api/admin/me` | 获取管理员信息 |

### 文件管理接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/files/upload` | 文件上传 |
| GET | `/api/files/download/:id` | 文件下载 |
| DELETE | `/api/files/:id` | 删除文件 |
| GET | `/api/files/list` | 文件列表 |

### 分享接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/share/create` | 创建分享链接 |
| GET | `/api/share/:token` | 访问分享内容 |
| DELETE | `/api/share/:id` | 删除分享 |

## 🏗️ 项目结构

```
HQyun/
├── backend/                 # 后端服务
│   ├── cmd/server/         # 服务入口
│   ├── internal/           # 内部模块
│   │   ├── api/           # API 控制器
│   │   ├── config/        # 配置管理
│   │   ├── database/      # 数据库操作
│   │   ├── model/         # 数据模型
│   │   └── storage/       # 存储管理
│   ├── public/            # 静态文件
│   └── configs/           # 配置文件
├── frontend/              # 前端资源
└── docker-compose.yml     # Docker 配置
```

## 🔒 安全特性

- JWT Token 认证
- 密码 bcrypt 加密
- 文件访问权限控制
- 分享链接权限验证
- XSS 和 CSRF 防护
- 文件类型安全检查

## 🌟 技术栈

### 后端
- **Go**: 高性能后端服务
- **Gin**: Web 框架
- **GORM**: ORM 数据库操作
- **SQLite**: 轻量级数据库
- **JWT**: 身份认证

### 前端
- **HTML5/CSS3**: 现代化界面
- **JavaScript**: 交互逻辑
- **响应式设计**: 多设备适配

## 📝 开发指南

### 本地开发

1. 克隆项目并进入目录
2. 安装 Go 依赖: `go mod tidy`
3. 运行开发服务器: `go run cmd/server/main.go`
4. 访问 `http://localhost:8080`

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 编写单元测试
- 提交前运行 `go vet` 检查

## 🤝 贡献指南

1. Fork 本项目
2. 创建特性分支: `git checkout -b feature/amazing-feature`
3. 提交更改: `git commit -m 'Add amazing feature'`
4. 推送分支: `git push origin feature/amazing-feature`
5. 提交 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

## 📞 联系方式

- 项目地址: [https://github.com/yourusername/HQyun](https://github.com/yourusername/HQyun)
- 问题反馈: [Issues](https://github.com/yourusername/HQyun/issues)

---

⭐ 如果这个项目对你有帮助，请给它一个 Star！