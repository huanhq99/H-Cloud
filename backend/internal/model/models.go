package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex;not null"`
	Password     string `gorm:"not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	Role         string `gorm:"default:user"` // admin 或 user
	StorageQuota int64  `gorm:"default:10737418240"` // 默认10GB，单位字节
	StorageUsed  int64  `gorm:"default:0"`
	LastLogin    time.Time
}

// Directory 目录模型
type Directory struct {
    gorm.Model
    Name        string `gorm:"not null"`
    Path        string `gorm:"not null;uniqueIndex:idx_user_path"`
    UserID      uint   `gorm:"index;uniqueIndex:idx_user_path"`
    ParentID    *uint  `gorm:"index"` // 父目录ID，根目录为nil
    IsMapping   bool   `gorm:"default:false"` // 是否为映射目录
    MappingPath string // 映射到本地的路径
    Files       []File `gorm:"foreignKey:DirectoryID"`
}

// File 文件模型
type File struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Path        string `gorm:"not null"` // 存储路径
	Size        int64  `gorm:"not null"` // 文件大小（字节）
	ContentType string // MIME类型
	Hash        string // 文件哈希，用于去重
	UserID      uint   `gorm:"index"`
	DirectoryID uint   `gorm:"index"`
	IsMapping   bool   `gorm:"default:false"` // 是否为映射文件
	MappingPath string // 映射到本地的路径
}

// Share 分享模型
type Share struct {
    gorm.Model
    UUID        string    `gorm:"uniqueIndex;not null"` // 分享的唯一标识
    UserID      uint      `gorm:"index"`
    FileID      *uint     // 分享的文件ID，如果是目录则为nil
    DirectoryID *uint     // 分享的目录ID，如果是文件则为nil
    ExpireAt    time.Time // 过期时间
    Password    string    // 访问密码，可为空
    ViewCount   int       `gorm:"default:0"` // 查看次数
    IsPublic    bool      `gorm:"default:true"` // 是否公开分享
    NoExpire    bool      `gorm:"default:false"` // 永久有效
}

// RecycleBin 回收站模型
type RecycleBin struct {
    gorm.Model
    UserID       uint      `gorm:"index;not null"`
    OriginalName string    `gorm:"not null"`           // 原始文件/目录名
    OriginalPath string    `gorm:"not null"`           // 原始路径
    StoragePath  string    `gorm:"not null"`           // 存储路径
    Size         int64     `gorm:"not null"`           // 文件大小（字节）
    ContentType  string                                // MIME类型
    ItemType     string    `gorm:"not null"`           // 类型：file 或 directory
    DeletedAt    time.Time `gorm:"not null"`           // 删除时间
    ExpireAt     time.Time `gorm:"not null"`           // 过期时间（30天后自动清理）
}