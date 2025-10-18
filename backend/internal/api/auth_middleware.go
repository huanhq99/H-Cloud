package api

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware 解析Authorization: Bearer <token> 并设置userID
func AuthMiddleware(secret string) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        // 如果已有 userID（例如开发模式注入），直接通过
        if _, ok := ctx.Get("userID"); ok {
            ctx.Next()
            return
        }

        auth := ctx.GetHeader("Authorization")
        if auth == "" || !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
            ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
            ctx.Abort()
            return
        }
        tokenStr := strings.TrimSpace(auth[len("Bearer "):])
        token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
            return []byte(secret), nil
        })
        if err != nil || !token.Valid {
            ctx.JSON(http.StatusUnauthorized, gin.H{"error": "令牌无效"})
            ctx.Abort()
            return
        }
        claims, ok := token.Claims.(*Claims)
        if !ok {
            ctx.JSON(http.StatusUnauthorized, gin.H{"error": "令牌解析失败"})
            ctx.Abort()
            return
        }
        ctx.Set("userID", claims.UserID)
        ctx.Set("role", claims.Role)
        ctx.Next()
    }
}