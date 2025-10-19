package database

import (
    "fmt"

    "github.com/huanhq99/H-Cloud/internal/config"
    "github.com/huanhq99/H-Cloud/internal/model"
    "gorm.io/driver/mysql"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "golang.org/x/crypto/bcrypt"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) (*gorm.DB, error) {
    var err error
    if cfg.Database.Host == "sqlite" {
        dbPath := cfg.Database.DBName
        if dbPath == "" {
            dbPath = "hqyun.db"
        }
        DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
    } else {
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
            cfg.Database.User,
            cfg.Database.Password,
            cfg.Database.Host,
            cfg.Database.Port,
            cfg.Database.DBName,
        )
        DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
    }
    if err != nil {
        return nil, err
    }

	// 自动迁移数据库模型
	err = migrateModels()
	if err != nil {
		return nil, err
	}

	// 种子默认管理员账号（如不存在）
	if err := seedDefaultAdmin(); err != nil {
		return nil, err
	}
	
	return DB, nil
}

// migrateModels 自动迁移数据库模型
func migrateModels() error {
    return DB.AutoMigrate(
        &model.User{},
        &model.File{},
        &model.Directory{},
        &model.Share{},
        &model.RecycleBin{},
    )
}

// seedDefaultAdmin 在首次启动时初始化一个默认管理员账户
func seedDefaultAdmin() error {
    var count int64
    if err := DB.Model(&model.User{}).Where("role = ?", "admin").Count(&count).Error; err != nil {
        return err
    }
    if count > 0 {
        return nil
    }
    // 没有管理员则创建默认管理员 admin / password
    // 使用 bcrypt 哈希存储密码
    hashed := []byte("password")
    // 尝试生成哈希（失败时仍以明文作为退路，但记录错误）
    if h, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost); err == nil {
        hashed = h
    }
    admin := &model.User{
        Username: "admin",
        Password: string(hashed),
        Email:    "admin@example.com",
        Role:     "admin",
    }
    // 若用户名或邮箱已被占用，则跳过
    var exist int64
    if err := DB.Model(&model.User{}).Where("username = ? OR email = ?", admin.Username, admin.Email).Count(&exist).Error; err != nil {
        return err
    }
    if exist > 0 {
        return nil
    }
    return DB.Create(admin).Error
}