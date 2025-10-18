package api

import (
    "encoding/hex"
    "net/http"
    "time"
    "crypto/rand"
    "fmt"

    "github.com/gin-gonic/gin"
    "github.com/huanhq99/H-Cloud/internal/model"
    "github.com/huanhq99/H-Cloud/internal/storage"
    "gorm.io/gorm"
)

// ShareController 分享控制器
type ShareController struct {
    DB *gorm.DB
}

// NewShareController 创建分享控制器
func NewShareController(db *gorm.DB) *ShareController {
    return &ShareController{DB: db}
}

// CreateShare 创建分享链接（支持文件或目录，当前仅文件下载）
func (c *ShareController) CreateShare(ctx *gin.Context) {
    var req struct {
        UserID      uint   `json:"userId"`      // 从请求中获取用户ID
        FileID      *uint  `json:"fileId"`
        DirectoryID *uint  `json:"directoryId"`
        ExpireHours int    `json:"expireHours"`
        ExpireDays  int    `json:"expireDays"`
        Forever     bool   `json:"forever"`
        Password    string `json:"password"`
        IsPublic    *bool  `json:"isPublic"`
    }
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
        return
    }

    // 使用请求中的用户ID
    userID := req.UserID
    if userID == 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
        return
    }

    if (req.FileID == nil && req.DirectoryID == nil) || (req.FileID != nil && req.DirectoryID != nil) {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "必须且仅选择文件或目录其中之一"})
        return
    }

    // 校验所有权
    if req.FileID != nil {
        var file model.File
        if err := c.DB.First(&file, *req.FileID).Error; err != nil {
            ctx.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
            return
        }
        if file.UserID != userID {
            ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限分享该文件"})
            return
        }
    }
    if req.DirectoryID != nil {
        var dir model.Directory
        if err := c.DB.First(&dir, *req.DirectoryID).Error; err != nil {
            ctx.JSON(http.StatusNotFound, gin.H{"error": "目录不存在"})
            return
        }
        if dir.UserID != userID {
            ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限分享该目录"})
            return
        }
    }

    // 生成UUID（安全随机16字节hex）
    b := make([]byte, 16)
    if _, err := rand.Read(b); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "生成分享ID失败"})
        return
    }
    uuid := hex.EncodeToString(b)

    // 过期时间与公开性
    var expireAt time.Time
    noExpire := false
    if req.Forever {
        noExpire = true
        expireAt = time.Now().AddDate(100, 0, 0) // 约定永久，设置一个很远的时间
    } else if req.ExpireDays > 0 {
        expireAt = time.Now().Add(time.Duration(req.ExpireDays) * 24 * time.Hour)
    } else {
        expireHours := req.ExpireHours
        if expireHours <= 0 { expireHours = 24 }
        expireAt = time.Now().Add(time.Duration(expireHours) * time.Hour)
    }
    isPublic := true
    if req.IsPublic != nil { isPublic = *req.IsPublic }

    share := model.Share{
        UUID:        uuid,
        UserID:      userID,
        FileID:      req.FileID,
        DirectoryID: req.DirectoryID,
        ExpireAt:    expireAt,
        NoExpire:    noExpire,
        Password:    req.Password,
        IsPublic:    isPublic,
    }
    if err := c.DB.Create(&share).Error; err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建分享失败"})
        return
    }

    // 返回相对链接，由前端拼接域名
    link := fmt.Sprintf("/api/shares/access/%s", uuid)
    ctx.JSON(http.StatusOK, gin.H{
        "message": "分享创建成功",
        "uuid": uuid,
        "link": link,
        "page": fmt.Sprintf("/share/%s", uuid),
        "expireAt": share.ExpireAt.Format(time.RFC3339),
        "isPermanent": share.NoExpire,
    })
}

// ListShares 列出我创建的分享
func (c *ShareController) ListShares(ctx *gin.Context) {
    userID, exists := ctx.Get("userID")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
        return
    }
    var shares []model.Share
    if err := c.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&shares).Error; err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取分享列表失败"})
        return
    }
    // 补充文件或目录名称
    resp := make([]gin.H, 0, len(shares))
    for _, s := range shares {
        item := gin.H{
            "uuid": s.UUID,
            "expireAt": s.ExpireAt.Format(time.RFC3339),
            "hasPassword": s.Password != "",
            "isPublic": s.IsPublic,
            "viewCount": s.ViewCount,
            "link": fmt.Sprintf("/api/shares/access/%s", s.UUID),
            "page": fmt.Sprintf("/share/%s", s.UUID),
            "isPermanent": s.NoExpire,
        }
        if s.FileID != nil {
            var f model.File
            if err := c.DB.First(&f, *s.FileID).Error; err == nil {
                item["type"] = "file"
                item["targetId"] = f.ID
                item["name"] = f.Name
                item["size"] = f.Size
                item["contentType"] = f.ContentType
            } else {
                item["type"] = "file"
                item["name"] = "(文件不存在)"
            }
        } else if s.DirectoryID != nil {
            var d model.Directory
            if err := c.DB.First(&d, *s.DirectoryID).Error; err == nil {
                item["type"] = "directory"
                item["targetId"] = d.ID
                item["name"] = d.Name
                item["path"] = d.Path
            } else {
                item["type"] = "directory"
                item["name"] = "(目录不存在)"
            }
        }
        resp = append(resp, item)
    }

    ctx.JSON(http.StatusOK, gin.H{"shares": resp})
}

// CheckShare 校验分享信息（用于前端判断是否需要密码）
func (c *ShareController) CheckShare(ctx *gin.Context) {
    uuid := ctx.Param("uuid")
    var share model.Share
    if err := c.DB.Where("uuid = ?", uuid).First(&share).Error; err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "分享不存在"})
        return
    }
    if !share.NoExpire && time.Now().After(share.ExpireAt) {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "分享已过期"})
        return
    }
    resp := gin.H{
        "uuid": share.UUID,
        "hasPassword": share.Password != "",
        "expireAt": share.ExpireAt.Format(time.RFC3339),
        "isPermanent": share.NoExpire,
        "link": fmt.Sprintf("/api/shares/access/%s", share.UUID),
    }
    if share.FileID != nil {
        var f model.File
        if err := c.DB.First(&f, *share.FileID).Error; err == nil {
            resp["type"] = "file"
            resp["name"] = f.Name
            resp["size"] = f.Size
            resp["contentType"] = f.ContentType
        } else {
            resp["type"] = "file"
            resp["name"] = "(文件不存在)"
        }
    } else if share.DirectoryID != nil {
        var d model.Directory
        if err := c.DB.First(&d, *share.DirectoryID).Error; err == nil {
            resp["type"] = "directory"
            resp["name"] = d.Name
            resp["path"] = d.Path
        } else {
            resp["type"] = "directory"
            resp["name"] = "(目录不存在)"
        }
    }
    ctx.JSON(http.StatusOK, resp)
}

// VerifyShare 验证分享密码（不返回文件，仅校验密码是否正确）
func (c *ShareController) VerifyShare(ctx *gin.Context) {
    uuid := ctx.Param("uuid")
    password := ctx.Query("password")

    var share model.Share
    if err := c.DB.Where("uuid = ?", uuid).First(&share).Error; err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "分享不存在"})
        return
    }
    if !share.NoExpire && time.Now().After(share.ExpireAt) {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "分享已过期"})
        return
    }
    if share.Password != "" && share.Password != password {
        ctx.JSON(http.StatusForbidden, gin.H{"error": "密码错误"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// RevokeShare 取消分享
func (c *ShareController) RevokeShare(ctx *gin.Context) {
    userID, exists := ctx.Get("userID")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
        return
    }
    uuid := ctx.Param("uuid")
    var share model.Share
    if err := c.DB.Where("uuid = ?", uuid).First(&share).Error; err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "分享不存在"})
        return
    }
    if share.UserID != userID.(uint) {
        ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限取消该分享"})
        return
    }
    if err := c.DB.Delete(&share).Error; err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "取消分享失败"})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"message": "分享已取消"})
}

// AccessShare 访问分享（下载或内联预览）
func (c *ShareController) AccessShare(ctx *gin.Context) {
    uuid := ctx.Param("uuid")
    password := ctx.Query("password")
    inline := ctx.Query("inline") == "1"

    var share model.Share
    if err := c.DB.Where("uuid = ?", uuid).First(&share).Error; err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "分享不存在"})
        return
    }
    if !share.NoExpire && time.Now().After(share.ExpireAt) {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "分享已过期"})
        return
    }
    if share.Password != "" && share.Password != password {
        ctx.JSON(http.StatusForbidden, gin.H{"error": "密码错误"})
        return
    }

    // 目前仅支持文件
    if share.FileID == nil {
        ctx.JSON(http.StatusNotImplemented, gin.H{"error": "暂未支持目录分享下载"})
        return
    }

    var fileRecord model.File
    if err := c.DB.First(&fileRecord, *share.FileID).Error; err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
        return
    }

    // 获取文件（使用文件所有者ID）
    f, err := storage.GetFile(fileRecord.UserID, fileRecord.Path)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件失败: " + err.Error()})
        return
    }
    defer f.Close()

    // 增加查看次数
    c.DB.Model(&share).UpdateColumn("view_count", gorm.Expr("view_count + 1"))

    if inline {
        ct := fileRecord.ContentType
        if ct == "" { ct = "application/octet-stream" }
        ctx.Header("Content-Type", ct)
        ctx.Header("Content-Length", fmt.Sprintf("%d", fileRecord.Size))
        
        // 直接复制文件内容而不是使用ctx.File，避免自动设置Content-Disposition
        http.ServeFile(ctx.Writer, ctx.Request, f.Name())
        return
    }

    ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileRecord.Name))
    ctx.Header("Content-Type", "application/octet-stream")
    ctx.Header("Content-Length", fmt.Sprintf("%d", fileRecord.Size))
    ctx.File(f.Name())
}