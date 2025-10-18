package security

import (
	"errors"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
)

// 允许的文件类型
var AllowedFileTypes = map[string][]string{
	"image": {".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg"},
	"document": {".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".md", ".rtf"},
	"video": {".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm"},
	"audio": {".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma"},
	"archive": {".zip", ".rar", ".7z", ".tar", ".gz", ".bz2"},
	"code": {".go", ".js", ".html", ".css", ".json", ".xml", ".yaml", ".yml", ".sql"},
}

// 危险文件扩展名黑名单
var DangerousExtensions = []string{
	".exe", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs", ".js", ".jar",
	".sh", ".php", ".asp", ".aspx", ".jsp", ".py", ".rb", ".pl",
}

// 最大文件大小配置 (字节)
const (
	MaxImageSize    = 10 * 1024 * 1024  // 10MB
	MaxDocumentSize = 50 * 1024 * 1024  // 50MB
	MaxVideoSize    = 500 * 1024 * 1024 // 500MB
	MaxAudioSize    = 100 * 1024 * 1024 // 100MB
	MaxArchiveSize  = 100 * 1024 * 1024 // 100MB
	MaxCodeSize     = 5 * 1024 * 1024   // 5MB
	MaxDefaultSize  = 20 * 1024 * 1024  // 20MB
)

// FileValidationResult 文件验证结果
type FileValidationResult struct {
	IsValid     bool
	FileType    string
	Extension   string
	ContentType string
	Error       error
}

// ValidateFileName 验证文件名安全性
func ValidateFileName(filename string) error {
	if filename == "" {
		return errors.New("文件名不能为空")
	}

	// 检查文件名长度
	if len(filename) > 255 {
		return errors.New("文件名过长")
	}

	// 检查危险字符
	dangerousChars := []string{"../", "..\\", "<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			return fmt.Errorf("文件名包含非法字符: %s", char)
		}
	}

	// 检查是否以点开头（隐藏文件）
	if strings.HasPrefix(filename, ".") {
		return errors.New("不允许上传隐藏文件")
	}

	return nil
}

// ValidateFilePath 验证文件路径安全性
func ValidateFilePath(path string) error {
	if path == "" {
		return errors.New("路径不能为空")
	}

	// 允许根目录
	if path == "/" {
		return nil
	}

	// 检查路径遍历攻击
	if strings.Contains(path, "../") || strings.Contains(path, "..\\") {
		return errors.New("检测到路径遍历攻击")
	}

	// 检查绝对路径（除了根目录）
	if filepath.IsAbs(path) {
		return errors.New("不允许使用绝对路径")
	}

	// 清理路径
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return errors.New("路径格式不正确")
	}

	return nil
}

// ValidateFileType 验证文件类型和大小
func ValidateFileType(filename string, size int64) FileValidationResult {
	result := FileValidationResult{
		IsValid: false,
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(filename))
	result.Extension = ext

	// 检查是否在危险扩展名列表中
	for _, dangerousExt := range DangerousExtensions {
		if ext == dangerousExt {
			result.Error = fmt.Errorf("不允许上传 %s 类型的文件", ext)
			return result
		}
	}

	// 确定文件类型和大小限制
	var maxSize int64 = MaxDefaultSize
	fileType := "other"

	for category, extensions := range AllowedFileTypes {
		for _, allowedExt := range extensions {
			if ext == allowedExt {
				fileType = category
				switch category {
				case "image":
					maxSize = MaxImageSize
				case "document":
					maxSize = MaxDocumentSize
				case "video":
					maxSize = MaxVideoSize
				case "audio":
					maxSize = MaxAudioSize
				case "archive":
					maxSize = MaxArchiveSize
				case "code":
					maxSize = MaxCodeSize
				}
				break
			}
		}
		if fileType != "other" {
			break
		}
	}

	result.FileType = fileType

	// 检查文件大小
	if size > maxSize {
		result.Error = fmt.Errorf("文件大小超过限制，最大允许 %d MB", maxSize/(1024*1024))
		return result
	}

	// 设置Content-Type
	result.ContentType = mime.TypeByExtension(ext)
	if result.ContentType == "" {
		result.ContentType = "application/octet-stream"
	}

	result.IsValid = true
	return result
}

// SanitizePath 清理和标准化路径
func SanitizePath(path string) string {
	// 移除开头的斜杠
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, "\\")

	// 清理路径
	path = filepath.Clean(path)

	// 如果是根目录，返回空字符串
	if path == "." {
		return ""
	}

	return path
}