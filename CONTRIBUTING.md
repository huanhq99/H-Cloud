# 贡献指南 🤝

感谢您对 H-Yun Cloud Drive 项目的关注！我们欢迎所有形式的贡献，包括但不限于代码、文档、测试、反馈和建议。

## 🌟 贡献方式

### 代码贡献
- 修复 Bug
- 添加新功能
- 性能优化
- 代码重构

### 文档贡献
- 改进文档
- 翻译文档
- 添加示例
- 修正错误

### 其他贡献
- 报告 Bug
- 提出功能建议
- 参与讨论
- 测试新版本

## 🚀 快速开始

### 1. Fork 项目

点击项目页面右上角的 "Fork" 按钮，将项目 Fork 到您的 GitHub 账户。

### 2. 克隆项目

```bash
git clone https://github.com/YOUR_USERNAME/HQyun.git
cd HQyun
```

### 3. 添加上游仓库

```bash
git remote add upstream https://github.com/ORIGINAL_OWNER/HQyun.git
```

### 4. 创建开发分支

```bash
git checkout -b feature/your-feature-name
# 或
git checkout -b fix/your-bug-fix
```

## 🛠️ 开发环境搭建

### 环境要求

- Go 1.19+
- Git
- 代码编辑器 (推荐 VS Code)

### 安装依赖

```bash
cd backend
go mod tidy
```

### 运行开发服务器

```bash
go run cmd/server/main.go
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/api

# 运行测试并显示覆盖率
go test -cover ./...
```

## 📝 代码规范

### Go 代码规范

我们遵循 Go 官方的代码规范和最佳实践：

#### 1. 格式化

使用 `gofmt` 格式化代码：

```bash
gofmt -w .
```

#### 2. 代码检查

使用 `go vet` 检查代码：

```bash
go vet ./...
```

#### 3. 静态分析

推荐使用 `golangci-lint`：

```bash
# 安装
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行检查
golangci-lint run
```

#### 4. 命名规范

- **包名**: 小写，简短，有意义
- **函数名**: 驼峰命名，公开函数首字母大写
- **变量名**: 驼峰命名，简洁明了
- **常量名**: 全大写，下划线分隔

```go
// 好的示例
package storage

const MaxFileSize = 100 * 1024 * 1024

type FileManager struct {
    basePath string
}

func (fm *FileManager) UploadFile(filename string) error {
    // 实现
}

// 不好的示例
package Storage  // 包名不应该大写

const max_file_size = 100 * 1024 * 1024  // 常量应该大写

func (fm *FileManager) upload_file(filename string) error {  // 应该使用驼峰命名
    // 实现
}
```

#### 5. 注释规范

- 公开的函数、类型、常量必须有注释
- 注释应该以被注释的名称开头
- 复杂的逻辑需要添加行内注释

```go
// FileManager 管理文件的上传、下载和删除操作
type FileManager struct {
    basePath string
}

// UploadFile 上传文件到指定路径
// 参数 filename 为文件名，返回错误信息
func (fm *FileManager) UploadFile(filename string) error {
    // 检查文件名是否合法
    if filename == "" {
        return errors.New("filename cannot be empty")
    }
    
    // 创建目标路径
    targetPath := filepath.Join(fm.basePath, filename)
    
    // 执行上传逻辑
    return fm.saveFile(targetPath)
}
```

#### 6. 错误处理

- 优先返回错误而不是 panic
- 使用有意义的错误信息
- 适当时使用错误包装

```go
// 好的示例
func (fm *FileManager) UploadFile(filename string) error {
    if filename == "" {
        return fmt.Errorf("filename cannot be empty")
    }
    
    if err := fm.validateFile(filename); err != nil {
        return fmt.Errorf("file validation failed: %w", err)
    }
    
    return nil
}

// 不好的示例
func (fm *FileManager) UploadFile(filename string) {
    if filename == "" {
        panic("filename cannot be empty")  // 不应该使用 panic
    }
}
```

### 前端代码规范

#### 1. HTML 规范

- 使用语义化标签
- 保持良好的缩进
- 添加必要的注释

```html
<!-- 好的示例 -->
<main class="container">
    <section class="file-list">
        <h2>文件列表</h2>
        <ul class="files">
            <li class="file-item">
                <span class="filename">document.pdf</span>
                <button class="download-btn">下载</button>
            </li>
        </ul>
    </section>
</main>
```

#### 2. CSS 规范

- 使用有意义的类名
- 保持样式的一致性
- 使用 CSS 变量定义主题色彩

```css
/* 好的示例 */
:root {
    --primary-color: #007bff;
    --secondary-color: #6c757d;
    --success-color: #28a745;
    --danger-color: #dc3545;
}

.file-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem;
    border-bottom: 1px solid var(--secondary-color);
}

.download-btn {
    background-color: var(--primary-color);
    color: white;
    border: none;
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    cursor: pointer;
}
```

#### 3. JavaScript 规范

- 使用现代 ES6+ 语法
- 保持函数简洁
- 添加适当的错误处理

```javascript
// 好的示例
class FileManager {
    constructor(apiBase) {
        this.apiBase = apiBase;
    }
    
    async uploadFile(file, path = '/') {
        try {
            const formData = new FormData();
            formData.append('file', file);
            formData.append('path', path);
            
            const response = await fetch(`${this.apiBase}/files/upload`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.getToken()}`
                },
                body: formData
            });
            
            if (!response.ok) {
                throw new Error(`Upload failed: ${response.statusText}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('Upload error:', error);
            throw error;
        }
    }
    
    getToken() {
        return localStorage.getItem('token');
    }
}
```

## 🧪 测试规范

### 单元测试

每个功能模块都应该有对应的单元测试：

```go
// file_manager_test.go
package storage

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestFileManager_UploadFile(t *testing.T) {
    tests := []struct {
        name     string
        filename string
        wantErr  bool
    }{
        {
            name:     "valid filename",
            filename: "test.txt",
            wantErr:  false,
        },
        {
            name:     "empty filename",
            filename: "",
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            fm := &FileManager{basePath: "/tmp"}
            err := fm.UploadFile(tt.filename)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 集成测试

测试 API 接口的完整流程：

```go
// api_test.go
package api

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestAuthController_Login(t *testing.T) {
    // 设置测试环境
    gin.SetMode(gin.TestMode)
    router := setupTestRouter()
    
    // 准备测试数据
    loginData := map[string]string{
        "username": "testuser",
        "password": "password123",
    }
    jsonData, _ := json.Marshal(loginData)
    
    // 创建请求
    req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    // 执行请求
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // 验证结果
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, float64(200), response["code"])
}
```

## 📋 提交规范

### Commit Message 格式

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### Type 类型

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

#### 示例

```bash
# 新功能
git commit -m "feat(auth): add admin login functionality"

# Bug 修复
git commit -m "fix(upload): resolve file size validation issue"

# 文档更新
git commit -m "docs: update API documentation"

# 代码重构
git commit -m "refactor(storage): improve file manager structure"
```

### Pull Request 规范

#### PR 标题

使用与 commit message 相同的格式：

```
feat(auth): add admin login functionality
```

#### PR 描述模板

```markdown
## 变更类型
- [ ] Bug 修复
- [ ] 新功能
- [ ] 代码重构
- [ ] 文档更新
- [ ] 其他

## 变更描述
简要描述本次变更的内容和目的。

## 测试
- [ ] 已添加单元测试
- [ ] 已添加集成测试
- [ ] 已手动测试
- [ ] 所有测试通过

## 检查清单
- [ ] 代码遵循项目规范
- [ ] 已更新相关文档
- [ ] 已添加必要的注释
- [ ] 没有引入新的警告

## 相关 Issue
Closes #123
```

## 🐛 Bug 报告

### 报告 Bug

使用 GitHub Issues 报告 Bug，请包含以下信息：

1. **Bug 描述**: 清楚地描述遇到的问题
2. **复现步骤**: 详细的复现步骤
3. **预期行为**: 描述您期望的正确行为
4. **实际行为**: 描述实际发生的情况
5. **环境信息**: 操作系统、Go 版本、浏览器等
6. **截图**: 如果适用，添加截图说明问题

### Bug 报告模板

```markdown
## Bug 描述
简要描述遇到的问题。

## 复现步骤
1. 打开应用
2. 点击 '...'
3. 输入 '...'
4. 看到错误

## 预期行为
描述您期望发生的情况。

## 实际行为
描述实际发生的情况。

## 环境信息
- 操作系统: [例如 macOS 12.0]
- Go 版本: [例如 1.21.0]
- 浏览器: [例如 Chrome 120.0]

## 附加信息
添加任何其他有助于解决问题的信息。
```

## 💡 功能建议

### 提出功能建议

使用 GitHub Issues 提出功能建议，请包含：

1. **功能描述**: 清楚地描述建议的功能
2. **使用场景**: 说明为什么需要这个功能
3. **解决方案**: 描述您认为的实现方式
4. **替代方案**: 考虑过的其他解决方案

### 功能建议模板

```markdown
## 功能描述
简要描述建议的功能。

## 问题背景
描述当前遇到的问题或限制。

## 建议的解决方案
详细描述您建议的实现方式。

## 替代方案
描述您考虑过的其他解决方案。

## 附加信息
添加任何其他相关信息，如截图、参考链接等。
```

## 🔄 开发流程

### 1. 选择 Issue

- 查看 [Issues](https://github.com/yourusername/HQyun/issues) 页面
- 选择标有 `good first issue` 的简单任务开始
- 在 Issue 下评论表明您要处理这个问题

### 2. 开发

```bash
# 同步上游代码
git fetch upstream
git checkout main
git merge upstream/main

# 创建功能分支
git checkout -b feature/issue-123

# 进行开发
# ... 编写代码 ...

# 提交代码
git add .
git commit -m "feat: add new feature for issue #123"
```

### 3. 测试

```bash
# 运行测试
go test ./...

# 检查代码质量
go vet ./...
golangci-lint run
```

### 4. 提交 PR

```bash
# 推送到您的 Fork
git push origin feature/issue-123

# 在 GitHub 上创建 Pull Request
```

### 5. 代码审查

- 响应审查意见
- 根据反馈修改代码
- 保持 PR 更新

### 6. 合并

- PR 通过审查后会被合并
- 删除功能分支

```bash
git checkout main
git pull upstream main
git branch -d feature/issue-123
```

## 📚 学习资源

### Go 语言学习

- [Go 官方文档](https://golang.org/doc/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://golang.org/doc/effective_go.html)

### Web 开发

- [Gin 框架文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [JWT 介绍](https://jwt.io/introduction/)

### 前端技术

- [MDN Web Docs](https://developer.mozilla.org/)
- [JavaScript 现代教程](https://zh.javascript.info/)
- [CSS Grid 指南](https://css-tricks.com/snippets/css/complete-guide-grid/)

## 🏆 贡献者

感谢所有为项目做出贡献的开发者！

<!-- 这里会自动生成贡献者列表 -->

## 📞 联系我们

- **GitHub Issues**: [项目 Issues](https://github.com/yourusername/HQyun/issues)
- **讨论区**: [GitHub Discussions](https://github.com/yourusername/HQyun/discussions)
- **邮箱**: your-email@example.com

## 📄 许可证

通过贡献代码，您同意您的贡献将在 [MIT 许可证](LICENSE) 下授权。

---

再次感谢您的贡献！🎉