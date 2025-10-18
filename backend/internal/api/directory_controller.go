package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/huanhq99/H-Cloud/internal/model"
	"github.com/huanhq99/H-Cloud/internal/storage"
	"github.com/huanhq99/H-Cloud/internal/security"
	"gorm.io/gorm"
)

// DirectoryController 目录控制器
type DirectoryController struct {
	DB *gorm.DB
}

// NewDirectoryController 创建目录控制器
func NewDirectoryController(db *gorm.DB) *DirectoryController {
	return &DirectoryController{DB: db}
}

// CreateDirectory 创建目录 - H-Yun盘版本
func (c *DirectoryController) CreateDirectory(ctx *gin.Context) {
	// H-Yun盘，使用固定用户ID
	userID := uint(1)

	// 获取请求参数
	var req struct {
		ParentPath string `json:"parentPath"`
		Name       string `json:"name" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 验证目录名安全性
	if err := security.ValidateFileName(req.Name); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "目录名不合法: " + err.Error()})
		return
	}

	// 如果没有指定父路径，默认为根目录
	if req.ParentPath == "" {
		req.ParentPath = "/"
	}

	// 验证父路径安全性
	if req.ParentPath != "/" {
		if err := security.ValidateFilePath(req.ParentPath); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "父路径不合法: " + err.Error()})
			return
		}
		req.ParentPath = security.SanitizePath(req.ParentPath)
	}

    // 预计算目标路径并检查重复
    var expectedPath string
    if req.ParentPath == "/" {
        expectedPath = req.Name
    } else {
        expectedPath = filepath.Join(req.ParentPath, req.Name)
    }

    var exist model.Directory
    if err := c.DB.Where("user_id = ? AND path = ?", userID, expectedPath).First(&exist).Error; err == nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "目录已存在"})
        return
    }

    // 创建目录（幂等，已存在则 ensureDir 不报错）
    dirPath, err := storage.CreateDirectory(userID, req.ParentPath, req.Name)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败: " + err.Error()})
        return
    }

	// 查找父目录ID
	var parentID *uint
	if req.ParentPath != "/" {
		var parentDir model.Directory
		result := c.DB.Where("user_id = ? AND path = ?", userID, req.ParentPath).First(&parentDir)
		if result.Error == nil {
			id := parentDir.ID
			parentID = &id
		}
	}

	// 创建目录记录
	directory := model.Directory{
		Name:     req.Name,
		Path:     dirPath,
		UserID:   userID,
		ParentID: parentID,
	}

    if err := c.DB.Create(&directory).Error; err != nil {
        // 复合唯一约束命中（sqlite / mysql 通用匹配）
        if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": "目录已存在"})
            return
        }
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "保存目录记录失败"})
        return
    }

	ctx.JSON(http.StatusOK, gin.H{
		"message": "目录创建成功",
		"directory": gin.H{
			"id":   directory.ID,
			"name": directory.Name,
			"path": directory.Path,
		},
	})
}

// MapDirectory 映射外部目录
func (c *DirectoryController) MapDirectory(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取请求参数
	var req struct {
		SourcePath string `json:"sourcePath" binding:"required"`
		TargetPath string `json:"targetPath" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

    // 先检查目标路径是否重复
    var exist model.Directory
    if err := c.DB.Where("user_id = ? AND path = ?", userID, req.TargetPath).First(&exist).Error; err == nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "目录已存在"})
        return
    }

    // 映射目录
    err := storage.MapDirectory(userID.(uint), req.SourcePath, req.TargetPath)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "映射目录失败: " + err.Error()})
        return
    }

	// 创建映射目录记录
	directory := model.Directory{
		Name:        filepath.Base(req.TargetPath),
		Path:        req.TargetPath,
		UserID:      userID.(uint),
		ParentID:    nil, // 映射目录默认放在根目录下
		IsMapping:   true,
		MappingPath: req.SourcePath,
	}

    if err := c.DB.Create(&directory).Error; err != nil {
        if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": "目录已存在"})
            return
        }
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "保存映射目录记录失败"})
        return
    }

	ctx.JSON(http.StatusOK, gin.H{
		"message": "目录映射成功",
		"directory": gin.H{
			"id":          directory.ID,
			"name":        directory.Name,
			"path":        directory.Path,
			"isMapping":   directory.IsMapping,
			"mappingPath": directory.MappingPath,
		},
	})
}

// ListDirectories 列出目录
func (c *DirectoryController) ListDirectories(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取父目录ID
	parentIDStr := ctx.DefaultQuery("parentId", "0")
	parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的父目录ID"})
		return
	}

    // 查询目录列表（兼容根目录 parent_id IS NULL）
    var directories []model.Directory
    var queryErr error
    if parentID == 0 {
        queryErr = c.DB.Where("user_id = ? AND parent_id IS NULL", userID).Find(&directories).Error
    } else {
        queryErr = c.DB.Where("user_id = ? AND parent_id = ?", userID, parentID).Find(&directories).Error
    }
    if queryErr != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取目录列表失败"})
        return
    }

	// 构建响应
	dirList := make([]gin.H, 0, len(directories))
	for _, dir := range directories {
		dirList = append(dirList, gin.H{
			"id":          dir.ID,
			"name":        dir.Name,
			"path":        dir.Path,
			"isMapping":   dir.IsMapping,
			"mappingPath": dir.MappingPath,
			"createdAt":   dir.CreatedAt.Format(time.RFC3339),
			"updatedAt":   dir.UpdatedAt.Format(time.RFC3339),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"directories": dirList,
	})
}

// DeleteDirectory 删除目录
func (c *DirectoryController) DeleteDirectory(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取目录ID
	dirID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的目录ID"})
		return
	}

	// 查询目录记录
	var directory model.Directory
	if err := c.DB.First(&directory, dirID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "目录不存在"})
		return
	}

	// 检查目录所有权
	if directory.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限删除此目录"})
		return
	}

	// 检查目录是否为空
	var fileCount int64
	c.DB.Model(&model.File{}).Where("directory_id = ?", dirID).Count(&fileCount)
	
	var subDirCount int64
	c.DB.Model(&model.Directory{}).Where("parent_id = ?", dirID).Count(&subDirCount)

	if fileCount > 0 || subDirCount > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "目录不为空，无法删除"})
		return
	}

	// 删除数据库中的目录记录
	if err := c.DB.Delete(&directory).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除目录记录失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "目录删除成功"})
}

// RenameDirectory 重命名目录 - H-Yun盘版本
func (c *DirectoryController) RenameDirectory(ctx *gin.Context) {
	// 获取目录路径参数
	dirPath := ctx.Query("path")
	if dirPath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请提供目录路径"})
		return
	}

	// 获取新目录名
	var req struct {
		NewName string `json:"newName" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请提供新目录名"})
		return
	}

	// 验证新目录名安全性
	if err := security.ValidateFileName(req.NewName); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "目录名不合法: " + err.Error()})
		return
	}

	// 构建新的目录路径
	parentDir := filepath.Dir(dirPath)
	newPath := filepath.Join(parentDir, req.NewName)

	// 直接使用存储路径，不再使用user_目录
	oldFullPath := filepath.Join(storage.StoragePath, dirPath)
	newFullPath := filepath.Join(storage.StoragePath, newPath)

	// 检查原目录是否存在
	if stat, err := os.Stat(oldFullPath); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "原目录不存在"})
		return
	} else if !stat.IsDir() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "指定路径不是目录"})
		return
	}

	// 检查新目录名是否已存在
	if _, err := os.Stat(newFullPath); err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "目标目录名已存在"})
		return
	}

	// 重命名目录
	if err := os.Rename(oldFullPath, newFullPath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "重命名失败: " + err.Error()})
		return
	}

	// 更新数据库记录（如果存在）
	var dirRecord model.Directory
	if err := c.DB.Where("path = ? AND user_id = ?", dirPath, 1).First(&dirRecord).Error; err == nil {
		// 更新目录记录的路径和名称
		dirRecord.Path = newPath
		dirRecord.Name = req.NewName
		c.DB.Save(&dirRecord)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "目录重命名成功",
		"directory": gin.H{
			"name": req.NewName,
			"path": newPath,
		},
	})
}