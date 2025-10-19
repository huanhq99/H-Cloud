package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	. "github.com/huanhq99/H-Cloud/internal/model"
)

type SearchController struct {
	DB *gorm.DB
}

func NewSearchController(db *gorm.DB) *SearchController {
	return &SearchController{DB: db}
}

// SearchFiles 搜索文件和目录
func (sc *SearchController) SearchFiles(c *gin.Context) {
	userIDStr := c.Query("userId")
	query := c.Query("query")
	
	if userIDStr == "" || query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID和搜索关键词不能为空"})
		return
	}
	
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}
	
	// 搜索文件
	var files []File
	err = sc.DB.Where("user_id = ? AND name LIKE ?", userID, "%"+query+"%").Find(&files).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索文件失败"})
		return
	}
	
	// 搜索目录
	var directories []Directory
	err = sc.DB.Where("user_id = ? AND name LIKE ?", userID, "%"+query+"%").Find(&directories).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索目录失败"})
		return
	}
	
	// 转换为统一格式
	var results []map[string]interface{}
	
	// 添加文件结果
	for _, file := range files {
		results = append(results, map[string]interface{}{
			"id":          file.ID,
			"name":        file.Name,
			"path":        file.Path,
			"size":        file.Size,
			"contentType": file.ContentType,
			"modTime":     file.UpdatedAt,
			"isDir":       false,
			"type":        "file",
		})
	}
	
	// 添加目录结果
	for _, dir := range directories {
		results = append(results, map[string]interface{}{
			"id":      dir.ID,
			"name":    dir.Name,
			"path":    dir.Path,
			"modTime": dir.CreatedAt,
			"isDir":   true,
			"type":    "directory",
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   len(results),
		"query":   query,
	})
}

// SearchByType 按类型搜索文件
func (sc *SearchController) SearchByType(c *gin.Context) {
	userIDStr := c.Query("userId")
	fileType := c.Query("type")
	
	if userIDStr == "" || fileType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID和文件类型不能为空"})
		return
	}
	
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}
	
	var files []File
	var whereClause string
	
	switch strings.ToLower(fileType) {
	case "image":
		whereClause = "user_id = ? AND content_type LIKE 'image/%'"
	case "video":
		whereClause = "user_id = ? AND content_type LIKE 'video/%'"
	case "audio":
		whereClause = "user_id = ? AND content_type LIKE 'audio/%'"
	case "document":
		whereClause = "user_id = ? AND (content_type LIKE 'application/pdf' OR content_type LIKE 'text/%' OR content_type LIKE 'application/msword%' OR content_type LIKE 'application/vnd.openxmlformats%')"
	case "archive":
		whereClause = "user_id = ? AND (content_type LIKE 'application/zip' OR content_type LIKE 'application/x-rar%' OR content_type LIKE 'application/x-tar%')"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型"})
		return
	}
	
	err = sc.DB.Where(whereClause, userID).Find(&files).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索文件失败"})
		return
	}
	
	// 转换为统一格式
	var results []map[string]interface{}
	for _, file := range files {
		results = append(results, map[string]interface{}{
			"id":          file.ID,
			"name":        file.Name,
			"path":        file.Path,
			"size":        file.Size,
			"contentType": file.ContentType,
			"modTime":     file.UpdatedAt,
			"isDir":       false,
			"type":        "file",
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   len(results),
		"type":    fileType,
	})
}