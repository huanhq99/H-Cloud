# H-Cloud Drive API 文档 📚

本文档详细描述了 H-Cloud Drive 的所有 API 接口。

## 📋 基本信息

- **Base URL**: `http://localhost:8080/api`
- **认证方式**: JWT Token
- **数据格式**: JSON
- **字符编码**: UTF-8

## 🔐 认证说明

### JWT Token 使用

大部分 API 需要在请求头中包含 JWT Token：

```http
Authorization: Bearer <your_jwt_token>
```

### Token 获取

通过用户登录或管理员登录接口获取 Token。

## 📝 通用响应格式

### 成功响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    // 具体数据
  }
}
```

### 错误响应

```json
{
  "code": 400,
  "message": "error message",
  "data": null
}
```

### 状态码说明

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 401 | 未授权/Token 无效 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 🔑 认证接口

### 用户注册

**POST** `/auth/register`

注册新用户账号。

#### 请求参数

```json
{
  "username": "string",     // 用户名 (3-20字符)
  "email": "string",        // 邮箱地址
  "password": "string"      // 密码 (6-50字符)
}
```

#### 响应示例

```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user_id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 用户登录

**POST** `/auth/login`

用户登录获取 Token。

#### 请求参数

```json
{
  "username": "string",     // 用户名或邮箱
  "password": "string"      // 密码
}
```

#### 响应示例

```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-01-02T00:00:00Z",
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com"
    }
  }
}
```

### 获取当前用户信息

**GET** `/auth/me`

获取当前登录用户的信息。

#### 请求头

```http
Authorization: Bearer <token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "storage_used": 1048576,
    "storage_limit": 1073741824
  }
}
```

## 👑 管理员接口

### 管理员登录

**POST** `/admin/login`

管理员登录获取管理员 Token。

#### 请求参数

```json
{
  "username": "string",     // 管理员用户名
  "password": "string"      // 管理员密码
}
```

#### 响应示例

```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-01-02T00:00:00Z",
    "admin": {
      "username": "admin"
    }
  }
}
```

### 管理员登出

**POST** `/admin/logout`

管理员登出，使 Token 失效。

#### 请求头

```http
Authorization: Bearer <admin_token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "登出成功",
  "data": null
}
```

### 获取管理员信息

**GET** `/admin/me`

获取当前管理员信息。

#### 请求头

```http
Authorization: Bearer <admin_token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "username": "admin",
    "login_time": "2024-01-01T12:00:00Z"
  }
}
```

## 📁 文件管理接口

### 文件上传

**POST** `/files/upload`

上传文件到指定目录。

#### 请求头

```http
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| file | File | 是 | 上传的文件 |
| path | string | 否 | 目标路径 (默认根目录) |

#### 响应示例

```json
{
  "code": 200,
  "message": "上传成功",
  "data": {
    "file_id": 123,
    "filename": "document.pdf",
    "size": 1048576,
    "path": "/documents/",
    "url": "/api/files/download/123",
    "uploaded_at": "2024-01-01T12:00:00Z"
  }
}
```

### 文件下载

**GET** `/files/download/:id`

下载指定文件。

#### 路径参数

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | int | 文件 ID |

#### 请求头

```http
Authorization: Bearer <token>
```

#### 响应

返回文件二进制数据，包含适当的 Content-Type 和 Content-Disposition 头。

### 获取文件列表

**GET** `/files/list`

获取指定目录下的文件列表。

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| path | string | 否 | 目录路径 (默认根目录) |
| page | int | 否 | 页码 (默认 1) |
| limit | int | 否 | 每页数量 (默认 20) |

#### 请求头

```http
Authorization: Bearer <token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "files": [
      {
        "id": 123,
        "filename": "document.pdf",
        "size": 1048576,
        "type": "file",
        "path": "/documents/",
        "created_at": "2024-01-01T12:00:00Z",
        "updated_at": "2024-01-01T12:00:00Z"
      },
      {
        "id": 124,
        "filename": "images",
        "size": 0,
        "type": "directory",
        "path": "/images/",
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z"
      }
    ],
    "total": 2,
    "page": 1,
    "limit": 20
  }
}
```

### 删除文件

**DELETE** `/files/:id`

删除指定文件或目录。

#### 路径参数

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | int | 文件/目录 ID |

#### 请求头

```http
Authorization: Bearer <token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "删除成功",
  "data": null
}
```

### 重命名文件

**PUT** `/files/:id/rename`

重命名文件或目录。

#### 路径参数

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | int | 文件/目录 ID |

#### 请求参数

```json
{
  "new_name": "string"      // 新文件名
}
```

#### 请求头

```http
Authorization: Bearer <token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "重命名成功",
  "data": {
    "id": 123,
    "old_name": "old_document.pdf",
    "new_name": "new_document.pdf"
  }
}
```

## 📂 目录管理接口

### 创建目录

**POST** `/directories/create`

在指定路径创建新目录。

#### 请求参数

```json
{
  "name": "string",         // 目录名称
  "path": "string"          // 父目录路径 (可选)
}
```

#### 请求头

```http
Authorization: Bearer <token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "目录创建成功",
  "data": {
    "id": 125,
    "name": "new_folder",
    "path": "/documents/new_folder/",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

### 获取目录树

**GET** `/directories/tree`

获取完整的目录树结构。

#### 请求头

```http
Authorization: Bearer <token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "tree": [
      {
        "id": 1,
        "name": "documents",
        "path": "/documents/",
        "type": "directory",
        "children": [
          {
            "id": 123,
            "name": "document.pdf",
            "path": "/documents/document.pdf",
            "type": "file",
            "size": 1048576
          }
        ]
      }
    ]
  }
}
```

## 🔗 分享接口

### 创建分享链接

**POST** `/share/create`

为文件或目录创建分享链接。

#### 请求参数

```json
{
  "file_id": 123,           // 文件/目录 ID
  "password": "string",     // 访问密码 (可选)
  "expires_at": "string",   // 过期时间 (可选, ISO 8601 格式)
  "allow_download": true    // 是否允许下载 (默认 true)
}
```

#### 请求头

```http
Authorization: Bearer <token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "分享链接创建成功",
  "data": {
    "share_id": "abc123def456",
    "share_url": "http://localhost:8080/share/abc123def456",
    "password": "1234",
    "expires_at": "2024-01-08T00:00:00Z",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

### 访问分享内容

**GET** `/share/:token`

访问分享的文件或目录。

#### 路径参数

| 参数名 | 类型 | 说明 |
|--------|------|------|
| token | string | 分享令牌 |

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| password | string | 否 | 访问密码 (如果设置了密码) |

#### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "file": {
      "id": 123,
      "filename": "document.pdf",
      "size": 1048576,
      "type": "file",
      "download_url": "/api/share/abc123def456/download"
    },
    "share_info": {
      "created_at": "2024-01-01T12:00:00Z",
      "expires_at": "2024-01-08T00:00:00Z",
      "allow_download": true
    }
  }
}
```

### 下载分享文件

**GET** `/share/:token/download`

下载分享的文件。

#### 路径参数

| 参数名 | 类型 | 说明 |
|--------|------|------|
| token | string | 分享令牌 |

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| password | string | 否 | 访问密码 |

#### 响应

返回文件二进制数据。

### 获取我的分享列表

**GET** `/share/my`

获取当前用户创建的所有分享链接。

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | int | 否 | 页码 (默认 1) |
| limit | int | 否 | 每页数量 (默认 20) |

#### 请求头

```http
Authorization: Bearer <token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "shares": [
      {
        "share_id": "abc123def456",
        "file_name": "document.pdf",
        "share_url": "http://localhost:8080/share/abc123def456",
        "has_password": true,
        "expires_at": "2024-01-08T00:00:00Z",
        "created_at": "2024-01-01T12:00:00Z",
        "view_count": 5
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 20
  }
}
```

### 删除分享链接

**DELETE** `/share/:id`

删除指定的分享链接。

#### 路径参数

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | 分享 ID |

#### 请求头

```http
Authorization: Bearer <token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "分享链接已删除",
  "data": null
}
```

## 🖼️ 图床接口

### 获取图片直链

**GET** `/images/direct/:id`

获取图片的直链地址。

#### 路径参数

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | int | 图片文件 ID |

#### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "direct_url": "http://localhost:8080/api/images/view/abc123def456",
    "markdown": "![image](http://localhost:8080/api/images/view/abc123def456)",
    "html": "<img src=\"http://localhost:8080/api/images/view/abc123def456\" alt=\"image\">",
    "bbcode": "[img]http://localhost:8080/api/images/view/abc123def456[/img]"
  }
}
```

### 查看图片

**GET** `/images/view/:token`

通过令牌查看图片。

#### 路径参数

| 参数名 | 类型 | 说明 |
|--------|------|------|
| token | string | 图片令牌 |

#### 响应

返回图片二进制数据。

## 📊 系统信息接口

### 获取系统信息

**GET** `/system/info`

获取系统基本信息。

#### 请求头

```http
Authorization: Bearer <admin_token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "version": "1.0.0",
    "uptime": "72h30m15s",
    "users_count": 10,
    "files_count": 1250,
    "storage_used": 1073741824,
    "storage_total": 10737418240,
    "system": {
      "os": "linux",
      "arch": "amd64",
      "go_version": "go1.21.0"
    }
  }
}
```

### 获取用户统计

**GET** `/system/users/stats`

获取用户统计信息。

#### 请求头

```http
Authorization: Bearer <admin_token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total_users": 10,
    "active_users": 8,
    "new_users_today": 2,
    "new_users_week": 5,
    "storage_usage": [
      {
        "user_id": 1,
        "username": "user1",
        "storage_used": 104857600,
        "files_count": 25
      }
    ]
  }
}
```

## 🔍 搜索接口

### 搜索文件

**GET** `/search`

搜索文件和目录。

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| q | string | 是 | 搜索关键词 |
| type | string | 否 | 文件类型 (file/directory/all) |
| page | int | 否 | 页码 (默认 1) |
| limit | int | 否 | 每页数量 (默认 20) |

#### 请求头

```http
Authorization: Bearer <token>
```

#### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "results": [
      {
        "id": 123,
        "filename": "document.pdf",
        "path": "/documents/document.pdf",
        "type": "file",
        "size": 1048576,
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 20,
    "query": "document"
  }
}
```

## 🚨 错误码参考

| 错误码 | 说明 | 解决方案 |
|--------|------|----------|
| 1001 | 用户名已存在 | 使用其他用户名 |
| 1002 | 邮箱已存在 | 使用其他邮箱 |
| 1003 | 用户名或密码错误 | 检查登录信息 |
| 1004 | Token 无效或过期 | 重新登录获取 Token |
| 2001 | 文件不存在 | 检查文件 ID |
| 2002 | 文件大小超限 | 压缩文件或分割上传 |
| 2003 | 文件类型不支持 | 检查允许的文件类型 |
| 2004 | 存储空间不足 | 清理文件或联系管理员 |
| 3001 | 分享链接不存在 | 检查分享链接 |
| 3002 | 分享链接已过期 | 联系分享者重新分享 |
| 3003 | 分享密码错误 | 输入正确的访问密码 |
| 4001 | 权限不足 | 联系管理员获取权限 |
| 5001 | 服务器内部错误 | 联系技术支持 |

## 📝 使用示例

### JavaScript 示例

```javascript
// 用户登录
async function login(username, password) {
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ username, password })
  });
  
  const result = await response.json();
  if (result.code === 200) {
    localStorage.setItem('token', result.data.token);
    return result.data;
  }
  throw new Error(result.message);
}

// 上传文件
async function uploadFile(file, path = '/') {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('path', path);
  
  const response = await fetch('/api/files/upload', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${localStorage.getItem('token')}`
    },
    body: formData
  });
  
  return await response.json();
}

// 获取文件列表
async function getFileList(path = '/', page = 1) {
  const response = await fetch(`/api/files/list?path=${encodeURIComponent(path)}&page=${page}`, {
    headers: {
      'Authorization': `Bearer ${localStorage.getItem('token')}`
    }
  });
  
  return await response.json();
}
```

### cURL 示例

```bash
# 用户登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'

# 上传文件
curl -X POST http://localhost:8080/api/files/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@/path/to/file.pdf" \
  -F "path=/documents/"

# 获取文件列表
curl -X GET "http://localhost:8080/api/files/list?path=/&page=1" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 创建分享链接
curl -X POST http://localhost:8080/api/share/create \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"file_id":123,"password":"1234","expires_at":"2024-01-08T00:00:00Z"}'
```

---

📞 **API 支持**: 如有疑问请查看 [Issues](https://github.com/huanhq99/H-Cloud/issues) 或提交问题。