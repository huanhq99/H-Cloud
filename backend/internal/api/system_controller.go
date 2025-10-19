package api

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huanhq99/H-Cloud/internal/storage"
	"gorm.io/gorm"
)

// SystemController 系统信息控制器
type SystemController struct {
	DB *gorm.DB
}

// NewSystemController 创建系统信息控制器
func NewSystemController(db *gorm.DB) *SystemController {
	return &SystemController{
		DB: db,
	}
}

// GetVersion 获取版本信息
func (sc *SystemController) GetVersion(c *gin.Context) {
	buildTime := time.Now().Format("2006-01-02")
	
	// 尝试从环境变量获取版本号
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "v1.2.4" // 默认版本号
	}
	
	versionInfo := gin.H{
		"version":    version,
		"build_time": buildTime,
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	}
	
	c.JSON(http.StatusOK, versionInfo)
}

// GetSystemInfo 获取系统信息
func (sc *SystemController) GetSystemInfo(ctx *gin.Context) {
	// 获取存储信息
	total, used, free, err := storage.GetSystemStorageInfo()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取存储信息失败: " + err.Error(),
		})
		return
	}

	// 获取用户数量
	var userCount int64
	sc.DB.Model(&struct {
		ID uint `gorm:"primarykey"`
	}{}).Count(&userCount)

	// 获取文件数量
	var fileCount int64
	sc.DB.Table("files").Count(&fileCount)

	ctx.JSON(http.StatusOK, gin.H{
		"version":     "v1.2.3",
		"build_time":  "2025-01-18",
		"go_version":  runtime.Version(),
		"os":          runtime.GOOS,
		"arch":        runtime.GOARCH,
		"uptime":      time.Since(time.Now().Add(-time.Hour * 24)).String(), // 示例运行时间
		"users_count": userCount,
		"files_count": fileCount,
		"storage": gin.H{
			"total": total,
			"used":  used,
			"free":  free,
		},
	})
}