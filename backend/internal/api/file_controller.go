package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"path/filepath"
	"strings"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/huanhq99/H-Cloud/internal/model"
	"github.com/huanhq99/H-Cloud/internal/storage"
	"github.com/huanhq99/H-Cloud/internal/security"
	"gorm.io/gorm"
)

// FileController 文件控制器
type FileController struct {
	DB *gorm.DB
}

// NewFileController 创建文件控制器
func NewFileController(db *gorm.DB) *FileController {
	return &FileController{DB: db}
}

// UploadFile 上传文件 - H-Yun盘版本
func (c *FileController) UploadFile(ctx *gin.Context) {
	// 获取上传的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件"})
		return
	}

	// 验证文件名安全性
	if err := security.ValidateFileName(file.Filename); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "文件名不合法: " + err.Error()})
		return
	}

	// 验证文件类型和大小
	validation := security.ValidateFileType(file.Filename, file.Size)
	if !validation.IsValid {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": validation.Error.Error()})
		return
	}

	// 获取目录路径，默认为根目录
	dirPath := ctx.DefaultPostForm("path", "/")
	
	// 验证路径安全性
	if err := security.ValidateFilePath(dirPath); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "路径不合法: " + err.Error()})
		return
	}

	// 清理路径
	dirPath = security.SanitizePath(dirPath)

	// 打开文件
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return
	}
	defer src.Close()

	// 保存文件到存储
	savedPath, err := storage.SaveFile(1, dirPath, file.Filename, file.Size, src)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败: " + err.Error()})
		return
	}

	// 使用验证结果中的Content-Type
	contentType := validation.ContentType

	// 保存文件信息到数据库
	fileModel := model.File{
		Name:        file.Filename,
		Path:        savedPath,
		Size:        file.Size,
		ContentType: contentType,
		UserID:      1, // 当前用户ID
	}

	if err := c.DB.Create(&fileModel).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件信息失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "文件上传成功",
		"file": gin.H{
			"id":          fileModel.ID,
			"name":        file.Filename,
			"path":        savedPath,
			"size":        file.Size,
			"contentType": contentType,
			"fileType":    validation.FileType,
		},
	})
}

// DownloadFile 下载文件
func (c *FileController) DownloadFile(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取文件ID
	fileID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	// 查询文件记录
	var fileRecord model.File
	if err := c.DB.First(&fileRecord, fileID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 检查文件所有权
	if fileRecord.UserID != userID.(uint) {
		// 检查是否有共享权限
		var share model.Share
		result := c.DB.Where("file_id = ? AND shared_with = ?", fileID, userID).First(&share)
		if result.Error != nil {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "没有权限访问此文件"})
			return
		}
	}

	// 获取文件
	file, err := storage.GetFile(userID.(uint), fileRecord.Path)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件失败: " + err.Error()})
		return
	}
	defer file.Close()

	// 设置响应头
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileRecord.Name))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Length", strconv.FormatInt(fileRecord.Size, 10))

	// 发送文件
	ctx.File(file.Name())
}

// DownloadFileByPath 基于路径下载文件 - H-Yun盘版本
func (c *FileController) DownloadFileByPath(ctx *gin.Context) {
	// 获取文件路径参数
	filePath := ctx.Query("path")
	if filePath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请提供文件路径"})
		return
	}

	// 获取用户ID参数（H-Yun盘版本）
	userIDStr := ctx.Query("userId")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请提供用户ID"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 查询文件记录
	var fileRecord model.File
	if err := c.DB.Where("path = ? AND user_id = ?", filePath, userID).First(&fileRecord).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 获取文件
	file, err := storage.GetFile(uint(userID), fileRecord.Path)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件失败: " + err.Error()})
		return
	}
	defer file.Close()

	// 设置响应头
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileRecord.Name))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Length", strconv.FormatInt(fileRecord.Size, 10))

	// 发送文件
	ctx.File(file.Name())
}

// ListFiles 列出文件 - H-Yun盘版本，直接读取映射路径
func (c *FileController) ListFiles(ctx *gin.Context) {
	// 获取路径参数，默认为根目录
	dirPath := ctx.DefaultQuery("path", "/")
	
	// 获取用户ID
	userIDStr := ctx.Query("userId")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户ID"})
		return
	}
	
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}
	
	// 使用物理目录扫描来获取所有文件和目录
	physicalFiles, err := storage.ListDirectory(uint(userID), dirPath)
	if err != nil {
		// 如果物理目录不存在，尝试从数据库获取（向后兼容）
		c.listFromDatabase(ctx, userID, dirPath)
		return
	}

	// 构建文件列表响应
	files := make([]gin.H, 0)
	directories := make([]gin.H, 0)
	
	// 创建数据库文件映射，用于获取额外信息
	dbFileMap := make(map[string]model.File)
	var dbFiles []model.File
	c.DB.Where("user_id = ?", userID).Find(&dbFiles)
	for _, file := range dbFiles {
		dbFileMap[file.Name] = file
	}

	// 处理物理目录中的每个条目
	for _, fileInfo := range physicalFiles {
		if fileInfo.IsDir() {
			// 目录
			directories = append(directories, gin.H{
				"id":        0, // 物理目录没有数据库ID
				"name":      fileInfo.Name(),
				"path":      filepath.Join(dirPath, fileInfo.Name()),
				"modTime":   fileInfo.ModTime().Format(time.RFC3339),
				"updatedAt": fileInfo.ModTime().Format(time.RFC3339),
				"isDir":     true,
			})
		} else {
			// 文件
			fileItem := gin.H{
				"name":        fileInfo.Name(),
				"path":        filepath.Join(dirPath, fileInfo.Name()),
				"size":        fileInfo.Size(),
				"modTime":     fileInfo.ModTime().Format(time.RFC3339),
				"updatedAt":   fileInfo.ModTime().Format(time.RFC3339),
				"isDir":       false,
			}
			
			// 如果数据库中有该文件的记录，使用数据库中的额外信息
			if dbFile, exists := dbFileMap[fileInfo.Name()]; exists {
				fileItem["id"] = dbFile.ID
				fileItem["contentType"] = dbFile.ContentType
			} else {
				fileItem["id"] = 0
				fileItem["contentType"] = "application/octet-stream"
			}
			
			files = append(files, fileItem)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"files":       files,
		"directories": directories,
	})
}

// listFromDatabase 从数据库获取文件列表（向后兼容方法）
func (c *FileController) listFromDatabase(ctx *gin.Context, userID int, dirPath string) {
	// 从数据库获取目录列表
	var dbDirectories []model.Directory
	var parentID *uint
	
	if dirPath == "/" {
		// 根目录，parent_id 为 NULL
		err := c.DB.Where("user_id = ? AND parent_id IS NULL", userID).Find(&dbDirectories).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取目录列表失败: " + err.Error()})
			return
		}
	} else {
		// 子目录，先找到父目录ID
		var parentDir model.Directory
		err := c.DB.Where("user_id = ? AND path = ?", userID, strings.TrimPrefix(dirPath, "/")).First(&parentDir).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "父目录不存在"})
			return
		}
		parentID = &parentDir.ID
		
		err = c.DB.Where("user_id = ? AND parent_id = ?", userID, parentID).Find(&dbDirectories).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取子目录列表失败: " + err.Error()})
			return
		}
	}
	
	// 从数据库获取文件列表
	var dbFiles []model.File
	if dirPath == "/" {
		// 根目录文件
		err := c.DB.Where("user_id = ? AND directory_id = 0", userID).Find(&dbFiles).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件列表失败: " + err.Error()})
			return
		}
	} else {
		// 子目录文件
		if parentID != nil {
			err := c.DB.Where("user_id = ? AND directory_id = ?", userID, *parentID).Find(&dbFiles).Error
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件列表失败: " + err.Error()})
				return
			}
		}
	}

	// 构建文件列表响应
	files := make([]gin.H, 0)
	directories := make([]gin.H, 0)

	// 添加数据库中的目录
	for _, dir := range dbDirectories {
		directories = append(directories, gin.H{
			"id":        dir.ID,
			"name":      dir.Name,
			"path":      "/" + dir.Path,
			"modTime":   dir.CreatedAt.Format(time.RFC3339),
			"updatedAt": dir.UpdatedAt.Format(time.RFC3339),
			"isDir":     true,
		})
	}
	
	// 添加数据库中的文件
	for _, file := range dbFiles {
		// 验证物理文件是否存在
		_, err := storage.GetFile(uint(userID), file.Path)
		if err != nil {
			// 物理文件不存在，从数据库中硬删除记录
			c.DB.Unscoped().Delete(&file)
			continue
		}
		
		files = append(files, gin.H{
			"id":          file.ID,
			"name":        file.Name,
			"path":        file.Path,
			"size":        file.Size,
			"contentType": file.ContentType,
			"modTime":     file.CreatedAt.Format(time.RFC3339),
			"updatedAt":   file.UpdatedAt.Format(time.RFC3339),
			"isDir":       false,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"files":       files,
		"directories": directories,
	})
}

// DeleteFile 删除文件 - H-Yun盘版本
func (c *FileController) DeleteFile(ctx *gin.Context) {
	// 获取文件路径参数
	filePath := ctx.Query("path")
	if filePath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请提供文件路径"})
		return
	}

	// 获取用户ID参数
	userIDStr := ctx.Query("userId")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请提供用户ID"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 查询文件记录
	var fileRecord model.File
	if err := c.DB.Where("path = ? AND user_id = ?", filePath, userID).First(&fileRecord).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 删除物理文件
	if err := storage.DeleteFile(uint(userID), filePath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除文件失败: " + err.Error()})
		return
	}

	// 删除数据库记录
	if err := c.DB.Delete(&fileRecord).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除数据库记录失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "文件删除成功"})
}

// RenameFile 重命名文件 - H-Yun盘版本
func (c *FileController) RenameFile(ctx *gin.Context) {
	// 获取文件路径参数
	filePath := ctx.Query("path")
	if filePath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请提供文件路径"})
		return
	}

	// 获取新文件名
	var req struct {
		NewName string `json:"newName" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请提供新文件名"})
		return
	}

	// 构建新的文件路径
	dir := filepath.Dir(filePath)
	newPath := filepath.Join(dir, req.NewName)

	// 获取用户存储路径
	userStoragePath := filepath.Join(storage.MappedPath, fmt.Sprintf("user_%d", 1))
	oldFullPath := filepath.Join(userStoragePath, filePath)
	newFullPath := filepath.Join(userStoragePath, newPath)

	// 重命名文件
	if err := os.Rename(oldFullPath, newFullPath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "重命名失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "文件重命名成功",
		"file": gin.H{
			"name": req.NewName,
			"path": newPath,
		},
	})
}

// GetImageDirect 图床功能：直接访问图片（无需认证）
func (c *FileController) GetImageDirect(ctx *gin.Context) {
	fileID := ctx.Param("id")
	id, err := strconv.ParseUint(fileID, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	// 查找文件记录
	var fileRecord model.File
	if err := c.DB.First(&fileRecord, uint(id)).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 检查是否为图片文件
	if !strings.HasPrefix(fileRecord.ContentType, "image/") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "该文件不是图片"})
		return
	}

	// 获取文件
	f, err := storage.GetFile(fileRecord.UserID, fileRecord.Path)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件失败: " + err.Error()})
		return
	}
	defer f.Close()

	// 设置响应头
	ctx.Header("Content-Type", fileRecord.ContentType)
	ctx.Header("Content-Length", fmt.Sprintf("%d", fileRecord.Size))
	ctx.Header("Cache-Control", "public, max-age=31536000") // 缓存1年
	ctx.Header("ETag", fmt.Sprintf(`"%d-%d"`, fileRecord.ID, fileRecord.UpdatedAt.Unix()))
	
	// 支持跨域访问（图床功能）
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type")

	// 返回文件内容
	ctx.DataFromReader(http.StatusOK, fileRecord.Size, fileRecord.ContentType, f, nil)
}