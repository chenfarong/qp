package middleware

import (
	"net/http"
	"strings"

	"zgame/internet/ssoauth/internal/util"

	"github.com/gin-gonic/gin"
)

// JWT 验证中间件
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization header中获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Authorization header is required"})
			c.Abort()
			return
		}

		// 检查Authorization header格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		// 解析JWT token
		tokenString := parts[1]
		claims, err := util.ParseJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("username", claims.Username)
		c.Next()
	}
}