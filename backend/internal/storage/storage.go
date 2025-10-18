package storage

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/huanhq99/H-Cloud/internal/config"
)

var (
	// StoragePath 存储路径
	StoragePath string
	// MappedPath 映射路径
	MappedPath string
)

// InitStorage 初始化存储服务
func InitStorage(cfg *config.Config) error {
	StoragePath = cfg.Storage.Path
	MappedPath = cfg.Storage.MappedPath

	// 确保存储目录存在
	if err := ensureDir(StoragePath); err != nil {
		return err
	}

	// 确保映射目录存在
	if err := ensureDir(MappedPath); err != nil {
		return err
	}

	return nil
}

// ensureDir 确保目录存在，如果不存在则创建
func ensureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// SaveFile 保存文件
func SaveFile(userID uint, dirPath string, filename string, fileSize int64, reader io.Reader) (string, error) {
	// 构建存储路径
	userDir := filepath.Join(StoragePath, fmt.Sprintf("user_%d", userID))
	if err := ensureDir(userDir); err != nil {
		return "", err
	}

	// 构建完整的目录路径
	fullDirPath := filepath.Join(userDir, dirPath)
	if err := ensureDir(fullDirPath); err != nil {
		return "", err
	}

	// 生成唯一文件名
	timestamp := time.Now().UnixNano()
	uniqueFilename := fmt.Sprintf("%d_%s", timestamp, filename)
	filePath := filepath.Join(fullDirPath, uniqueFilename)

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 写入文件内容
	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(filePath) // 如果写入失败，删除文件
		return "", err
	}

	// 返回相对于用户目录的路径
	relPath, err := filepath.Rel(userDir, filePath)
	if err != nil {
		return "", err
	}

	return relPath, nil
}

// GetFile 获取文件
func GetFile(userID uint, filePath string) (*os.File, error) {
	// 构建完整的文件路径
	fullPath := filepath.Join(StoragePath, fmt.Sprintf("user_%d", userID), filePath)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, errors.New("文件不存在")
	}

	// 打开文件
	return os.Open(fullPath)
}

// DeleteFile 删除文件
func DeleteFile(userID uint, filePath string) error {
	// 构建完整的文件路径
	fullPath := filepath.Join(StoragePath, fmt.Sprintf("user_%d", userID), filePath)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return errors.New("文件不存在")
	}

	// 删除文件
	return os.Remove(fullPath)
}

// CreateDirectory 创建目录
func CreateDirectory(userID uint, dirPath string, dirName string) (string, error) {
	// 构建用户目录
	userDir := filepath.Join(StoragePath, fmt.Sprintf("user_%d", userID))
	if err := ensureDir(userDir); err != nil {
		return "", err
	}

	// 构建完整的目录路径
	fullPath := filepath.Join(userDir, dirPath, dirName)
	if err := ensureDir(fullPath); err != nil {
		return "", err
	}

	// 返回相对于用户目录的路径
	relPath, err := filepath.Rel(userDir, fullPath)
	if err != nil {
		return "", err
	}

	return relPath, nil
}

// MapDirectory 映射目录
func MapDirectory(userID uint, sourcePath string, targetPath string) error {
	// 检查源目录是否存在
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return errors.New("源目录不存在")
	}

	// 构建目标目录路径
	userMappedDir := filepath.Join(MappedPath, fmt.Sprintf("user_%d", userID))
	if err := ensureDir(userMappedDir); err != nil {
		return err
	}

	// 构建完整的目标路径
	fullTargetPath := filepath.Join(userMappedDir, targetPath)
	
	// 创建符号链接
	if err := os.Symlink(sourcePath, fullTargetPath); err != nil {
		return err
	}

	return nil
}

// ListDirectory 列出目录内容
func ListDirectory(userID uint, dirPath string) ([]fs.FileInfo, error) {
	// 构建完整的目录路径
	fullPath := filepath.Join(StoragePath, fmt.Sprintf("user_%d", userID), dirPath)

	// 检查目录是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, errors.New("目录不存在")
	}

	// 读取目录内容
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	
	// 转换为FileInfo切片
	fileInfos := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, info)
	}
	
	return fileInfos, nil
}

// GetSystemStorageInfo 获取系统存储信息
func GetSystemStorageInfo() (total int64, used int64, free int64, err error) {
	var stat syscall.Statfs_t
	err = syscall.Statfs(StoragePath, &stat)
	if err != nil {
		return 0, 0, 0, err
	}

	// 计算总容量、已用容量和可用容量
	total = int64(stat.Blocks) * int64(stat.Bsize)
	free = int64(stat.Bavail) * int64(stat.Bsize)
	used = total - free

	return total, used, free, nil
}