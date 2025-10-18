package api

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/huanhq/hqyun/internal/config"
    "github.com/huanhq/hqyun/internal/model"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

// Claims 自定义JWT声明
type Claims struct {
    UserID uint   `json:"uid"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

// AuthController 账户控制器
type AuthController struct {
    DB        *gorm.DB
    JWTSecret string
    ExpiresIn int // 小时
}

func NewAuthController(db *gorm.DB, cfg *config.Config) *AuthController {
    return &AuthController{DB: db, JWTSecret: cfg.JWT.Secret, ExpiresIn: cfg.JWT.ExpiresIn}
}

// generateToken 生成JWT令牌
func (a *AuthController) generateToken(uid uint, role string) (string, error) {
    claims := &Claims{
        UserID: uid,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(a.ExpiresIn) * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(a.JWTSecret))
}

// Register 用户注册
func (a *AuthController) Register(ctx *gin.Context) {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
        Email    string `json:"email"`
    }
    if err := ctx.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" || req.Email == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的注册参数"})
        return
    }

    // 唯一性检查
    var cnt int64
    if err := a.DB.Model(&model.User{}).Where("username = ? OR email = ?", req.Username, req.Email).Count(&cnt).Error; err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "检查用户失败"})
        return
    }
    if cnt > 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "用户名或邮箱已存在"})
        return
    }

    // 加密密码
    hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "加密密码失败"})
        return
    }

    u := &model.User{Username: req.Username, Password: string(hashed), Email: req.Email, Role: "user"}
    if err := a.DB.Create(u).Error; err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
        return
    }

    token, err := a.generateToken(u.ID, u.Role)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"token": token, "user": gin.H{"id": u.ID, "username": u.Username, "email": u.Email, "role": u.Role}})
}

// Login 用户登录（支持用户名或邮箱）
func (a *AuthController) Login(ctx *gin.Context) {
    var req struct {
        Username string `json:"username"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := ctx.ShouldBindJSON(&req); err != nil || req.Password == "" || (req.Username == "" && req.Email == "") {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的登录参数"})
        return
    }

    var u model.User
    var q *gorm.DB
    if req.Email != "" { q = a.DB.Where("email = ?", req.Email) } else { q = a.DB.Where("username = ?", req.Username) }
    if err := q.First(&u).Error; err != nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
        return
    }
    if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
        return
    }
    a.DB.Model(&u).Update("last_login", time.Now())
    token, err := a.generateToken(u.ID, u.Role)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"token": token, "user": gin.H{"id": u.ID, "username": u.Username, "email": u.Email, "role": u.Role}})
}

// Me 当前用户信息
func (a *AuthController) Me(ctx *gin.Context) {
    uidVal, exists := ctx.Get("userID")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
        return
    }
    var u model.User
    if err := a.DB.First(&u, uidVal.(uint)).Error; err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"user": gin.H{"id": u.ID, "username": u.Username, "email": u.Email, "role": u.Role}})
}

// UpdateEmail 更新邮箱
func (a *AuthController) UpdateEmail(ctx *gin.Context) {
    uidVal, exists := ctx.Get("userID")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
        return
    }
    var req struct{ Email string `json:"email"` }
    if err := ctx.ShouldBindJSON(&req); err != nil || req.Email == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的邮箱"})
        return
    }
    // 唯一性检查
    var cnt int64
    if err := a.DB.Model(&model.User{}).Where("email = ?", req.Email).Count(&cnt).Error; err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "检查邮箱失败"})
        return
    }
    if cnt > 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "邮箱已被使用"})
        return
    }
    if err := a.DB.Model(&model.User{}).Where("id = ?", uidVal.(uint)).Update("email", req.Email).Error; err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新邮箱失败"})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"message": "邮箱已更新"})
}