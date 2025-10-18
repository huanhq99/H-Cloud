package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huanhq/hqyun/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// AdminController 管理员控制器
type AdminController struct {
	Config *config.Config
}

// NewAdminController 创建管理员控制器
func NewAdminController(cfg *config.Config) *AdminController {
	return &AdminController{
		Config: cfg,
	}
}

// AdminClaims 管理员JWT声明
type AdminClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	Username  string `json:"username"`
}

// Login 管理员登录
func (ac *AdminController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 验证用户名和密码
	if req.Username != ac.Config.Admin.Username || req.Password != ac.Config.Admin.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 生成JWT token
	expiresAt := time.Now().Add(time.Duration(ac.Config.JWT.ExpiresIn) * time.Hour)
	claims := AdminClaims{
		Username: req.Username,
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "hqyun-admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(ac.Config.JWT.Secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt.Unix(),
		Username:  req.Username,
	})
}

// Logout 管理员登出
func (ac *AdminController) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// Me 获取当前管理员信息
func (ac *AdminController) Me(c *gin.Context) {
	// 从中间件获取用户信息
	username, exists := c.Get("admin_username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username": username,
		"role":     "admin",
	})
}

// AdminAuthMiddleware 管理员认证中间件
func AdminAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取token
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证token"})
			c.Abort()
			return
		}

		// 移除Bearer前缀
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// 解析token
		token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			c.Abort()
			return
		}

		// 获取claims
		if claims, ok := token.Claims.(*AdminClaims); ok {
			// 检查是否为管理员角色
			if claims.Role != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
				c.Abort()
				return
			}

			// 将用户信息存储到上下文
			c.Set("admin_username", claims.Username)
			c.Set("admin_role", claims.Role)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token格式"})
			c.Abort()
			return
		}

		c.Next()
	}
}