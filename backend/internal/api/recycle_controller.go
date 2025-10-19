package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huanhq99/H-Cloud/internal/model"
	"github.com/huanhq99/H-Cloud/internal/storage"
	"gorm.io/gorm"
)

// RecycleController 回收站控制器
type RecycleController struct {
	DB *gorm.DB
}

// NewRecycleController 创建回收站控制器
func NewRecycleController(db *gorm.DB) *RecycleController {
	return &RecycleController{DB: db}
}

// ListRecycleBin 获取回收站列表
func (c *RecycleController) ListRecycleBin(ctx *gin.Context) {
	// 获取用户ID
	userIDStr := ctx.Query("userId")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户ID"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 查询回收站中的文件
	var recycleBinItems []model.RecycleBin
	if err := c.DB.Where("user_id = ?", userID).Order("deleted_at DESC").Find(&recycleBinItems).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取回收站列表失败: " + err.Error()})
		return
	}

	// 构建响应数据
	items := make([]gin.H, 0)
	for _, item := range recycleBinItems {
		items = append(items, gin.H{
			"id":           item.ID,
			"originalName": item.OriginalName,
			"originalPath": item.OriginalPath,
			"size":         item.Size,
			"contentType":  item.ContentType,
			"itemType":     item.ItemType,
			"deletedAt":    item.DeletedAt.Format(time.RFC3339),
			"expireAt":     item.ExpireAt.Format(time.RFC3339),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"items": items})
}

// RestoreFromRecycleBin 从回收站恢复文件
func (c *RecycleController) RestoreFromRecycleBin(ctx *gin.Context) {
	// 获取回收站项目ID
	itemID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 获取用户ID
	userIDStr := ctx.Query("userId")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户ID"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 查询回收站项目
	var recycleBinItem model.RecycleBin
	if err := c.DB.Where("id = ? AND user_id = ?", itemID, userID).First(&recycleBinItem).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "回收站项目不存在"})
		return
	}

	// 检查原始位置是否已存在同名文件
	originalFullPath := filepath.Join(storage.StoragePath, recycleBinItem.OriginalPath)
	if _, err := os.Stat(originalFullPath); err == nil {
		// 文件已存在，生成新名称
		ext := filepath.Ext(recycleBinItem.OriginalName)
		nameWithoutExt := recycleBinItem.OriginalName[:len(recycleBinItem.OriginalName)-len(ext)]
		counter := 1
		for {
			newName := nameWithoutExt + "_恢复" + strconv.Itoa(counter) + ext
			newPath := filepath.Join(filepath.Dir(originalFullPath), newName)
			if _, err := os.Stat(newPath); os.IsNotExist(err) {
				originalFullPath = newPath
				recycleBinItem.OriginalName = newName
				recycleBinItem.OriginalPath = filepath.Join(filepath.Dir(recycleBinItem.OriginalPath), newName)
				break
			}
			counter++
		}
	}

	// 移动文件从回收站存储位置到原始位置
	recycleStoragePath := filepath.Join(storage.StoragePath, ".recycle", recycleBinItem.StoragePath)
	if err := os.Rename(recycleStoragePath, originalFullPath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "恢复文件失败: " + err.Error()})
		return
	}

	// 如果是文件，恢复到数据库
	if recycleBinItem.ItemType == "file" {
		fileModel := model.File{
			Name:        recycleBinItem.OriginalName,
			Path:        recycleBinItem.OriginalPath,
			Size:        recycleBinItem.Size,
			ContentType: recycleBinItem.ContentType,
			UserID:      uint(userID),
		}
		if err := c.DB.Create(&fileModel).Error; err != nil {
			// 如果数据库恢复失败，回滚文件移动
			os.Rename(originalFullPath, recycleStoragePath)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "恢复文件记录失败: " + err.Error()})
			return
		}
	}

	// 从回收站中删除记录
	if err := c.DB.Delete(&recycleBinItem).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除回收站记录失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "文件恢复成功"})
}

// PermanentDelete 永久删除回收站中的文件
func (c *RecycleController) PermanentDelete(ctx *gin.Context) {
	// 获取回收站项目ID
	itemID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 获取用户ID
	userIDStr := ctx.Query("userId")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户ID"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 查询回收站项目
	var recycleBinItem model.RecycleBin
	if err := c.DB.Where("id = ? AND user_id = ?", itemID, userID).First(&recycleBinItem).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "回收站项目不存在"})
		return
	}

	// 删除物理文件
	recycleStoragePath := filepath.Join(storage.StoragePath, ".recycle", recycleBinItem.StoragePath)
	if err := os.RemoveAll(recycleStoragePath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除物理文件失败: " + err.Error()})
		return
	}

	// 从回收站中删除记录
	if err := c.DB.Delete(&recycleBinItem).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除回收站记录失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "文件已永久删除"})
}

// EmptyRecycleBin 清空回收站
func (c *RecycleController) EmptyRecycleBin(ctx *gin.Context) {
	// 获取用户ID
	userIDStr := ctx.Query("userId")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户ID"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 查询用户的所有回收站项目
	var recycleBinItems []model.RecycleBin
	if err := c.DB.Where("user_id = ?", userID).Find(&recycleBinItems).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取回收站列表失败: " + err.Error()})
		return
	}

	// 删除所有物理文件
	for _, item := range recycleBinItems {
		recycleStoragePath := filepath.Join(storage.StoragePath, ".recycle", item.StoragePath)
		os.RemoveAll(recycleStoragePath) // 忽略错误，继续删除其他文件
	}

	// 删除所有回收站记录
	if err := c.DB.Where("user_id = ?", userID).Delete(&model.RecycleBin{}).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "清空回收站记录失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "回收站已清空"})
}

// CleanExpiredItems 清理过期的回收站项目（定时任务调用）
func (c *RecycleController) CleanExpiredItems() error {
	// 查询所有过期的项目
	var expiredItems []model.RecycleBin
	if err := c.DB.Where("expire_at < ?", time.Now()).Find(&expiredItems).Error; err != nil {
		return err
	}

	// 删除过期的物理文件
	for _, item := range expiredItems {
		recycleStoragePath := filepath.Join(storage.StoragePath, ".recycle", item.StoragePath)
		os.RemoveAll(recycleStoragePath) // 忽略错误
	}

	// 删除过期的数据库记录
	return c.DB.Where("expire_at < ?", time.Now()).Delete(&model.RecycleBin{}).Error
}