package api

import (
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/huanhq99/H-Cloud/internal/config"
    "github.com/huanhq99/H-Cloud/internal/model"
    "github.com/huanhq99/H-Cloud/internal/storage"
    "gorm.io/gorm"
)

// isImageFile 检查文件是否为图片类型
func isImageFile(filename string) bool {
    ext := strings.ToLower(filename)
    return strings.HasSuffix(ext, ".jpg") || strings.HasSuffix(ext, ".jpeg") || 
           strings.HasSuffix(ext, ".png") || strings.HasSuffix(ext, ".gif") || 
           strings.HasSuffix(ext, ".bmp") || strings.HasSuffix(ext, ".webp") || 
           strings.HasSuffix(ext, ".svg")
}

// SetupRouter 设置路由
func SetupRouter(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
    // 健康检查
    r.GET("/health", func(ctx *gin.Context) {
        ctx.JSON(200, gin.H{"status": "ok"})
    })

    // 管理界面 - 根路径、/api 和 /api.html 都可以访问
    r.Use(func(c *gin.Context) {
        // 禁用缓存，确保文件更新能立即生效
        c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
        c.Header("Pragma", "no-cache")
        c.Header("Expires", "0")
        c.Next()
    })
    r.StaticFile("/", "./public/api.html")
    r.StaticFile("/api", "./public/api.html")
    r.StaticFile("/api.html", "./public/api.html")
    r.StaticFile("/login.html", "./public/login.html")  // 添加登录页面路由
    
    // 加载HTML模板
    r.LoadHTMLGlob("public/*.html")
    
    // 分享路由 - 智能处理图片和其他文件
    r.GET("/share/:uuid", func(ctx *gin.Context) {
        uuid := ctx.Param("uuid")
        
        // 检查是否为图片文件
        var share model.Share
        if err := db.Where("uuid = ?", uuid).First(&share).Error; err != nil {
            ctx.File("./public/share.html")
            return
        }
        
        // 如果是文件分享且为图片文件，都显示分享页面而不是直接重定向
        if share.FileID != nil {
            var fileRecord model.File
            if err := db.First(&fileRecord, *share.FileID).Error; err == nil {
                if isImageFile(fileRecord.Name) {
                    // 所有图片都显示分享页面，让JavaScript处理图片预览
                    ctx.File("./public/share.html")
                    return
                }
            }
        }
        
        // 其他情况显示分享页面
        ctx.File("./public/share.html")
    })

    // 创建控制器实例
    authController := NewAuthController(db, cfg)
    fileController := NewFileController(db)
    dirController := NewDirectoryController(db)
    shareController := NewShareController(db)
    adminController := NewAdminController(cfg)
    systemController := NewSystemController(db)

    // API 路由组
    api := r.Group("/api")
    {
        // 账户相关路由
        api.POST("/auth/register", authController.Register)
        api.POST("/auth/login", authController.Login)
        auth := api.Group("/auth")
        auth.Use(AuthMiddleware(cfg.JWT.Secret))
        {
            auth.GET("/me", authController.Me)
            auth.PUT("/email", authController.UpdateEmail)
        }

        // 文件相关路由 - 移除认证中间件，实现H-Yun盘
        files := api.Group("/files")
        {
            files.POST("/upload", fileController.UploadFile)
            files.GET("/download/:id", fileController.DownloadFile)
            files.GET("/download", fileController.DownloadFileByPath)  // 新增：基于路径的下载
            files.GET("/list", fileController.ListFiles)
            files.DELETE("/delete", fileController.DeleteFile)
		files.PUT("/rename", fileController.RenameFile)
        }

        // 图床功能路由 - 无需认证的图片直链访问
        api.GET("/image/:id", fileController.GetImageDirect)

        // 目录相关路由
        dirs := api.Group("/directories")
        {
            dirs.POST("/create", dirController.CreateDirectory)
            dirs.POST("/map", dirController.MapDirectory)
            dirs.GET("/list", dirController.ListDirectories)
            dirs.DELETE("/:id", dirController.DeleteDirectory)
            dirs.PUT("/rename", dirController.RenameDirectory)
        }

        // 分享相关路由 - 移除认证，实现H-Yun盘
        shares := api.Group("/shares")
        {
            // 公开访问的接口
            shares.GET("/check/:uuid", shareController.CheckShare)
            shares.GET("/verify/:uuid", shareController.VerifyShare)
            shares.GET("/access/:uuid", shareController.AccessShare)

            // 分享管理接口 - 移除认证
            shares.POST("/create", shareController.CreateShare)
            shares.GET("/list", shareController.ListShares)
            shares.DELETE("/:uuid", shareController.RevokeShare)
        }

        // 系统信息路由
        api.GET("/system/storage", GetSystemStorageInfo)
        api.GET("/version", systemController.GetVersion)
        api.GET("/system/info", systemController.GetSystemInfo)

        // 管理员相关路由
        api.POST("/admin/login", adminController.Login)
        admin := api.Group("/admin")
        admin.Use(AdminAuthMiddleware(cfg.JWT.Secret))
        {
            admin.POST("/logout", adminController.Logout)
            admin.GET("/me", adminController.Me)
        }
    }
}

// GetSystemStorageInfo 获取系统存储信息
func GetSystemStorageInfo(ctx *gin.Context) {
	total, used, free, err := storage.GetSystemStorageInfo()
	if err != nil {
		ctx.JSON(500, gin.H{"error": "获取存储信息失败: " + err.Error()})
		return
	}

	ctx.JSON(200, gin.H{
		"total": total,
		"used":  used,
		"free":  free,
	})
}