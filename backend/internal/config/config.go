package config

import (
    "strings"
    "github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Storage  StorageConfig  `mapstructure:"storage"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Admin    AdminConfig    `mapstructure:"admin"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Path       string `mapstructure:"path"`
	MappedPath string `mapstructure:"mapped_path"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret    string `mapstructure:"secret"`
	ExpiresIn int    `mapstructure:"expires_in"` // 过期时间（小时）
}

// AdminConfig 管理员配置
type AdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./configs")
    viper.AddConfigPath(".")

    // 设置默认值
    setDefaults()

    // 从环境变量读取，支持嵌套键（. 转换为 _）
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，使用默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults 设置默认配置
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.port", 8080)

	// 数据库默认配置
	viper.SetDefault("database.host", "mysql")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.user", "hqyun")
	viper.SetDefault("database.password", "hqyun_password")
	viper.SetDefault("database.dbname", "hqyun")

	// 存储默认配置
	viper.SetDefault("storage.path", "/data/storage")
	viper.SetDefault("storage.mapped_path", "/data/mapped_storage")

	// JWT默认配置
	viper.SetDefault("jwt.secret", "hqyun_secret_key")
	viper.SetDefault("jwt.expires_in", 24) // 24小时

	// 管理员默认配置
	viper.SetDefault("admin.username", "admin")
	viper.SetDefault("admin.password", "password")
}