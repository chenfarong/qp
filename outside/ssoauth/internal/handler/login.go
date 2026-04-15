package handler

import (
	"net/http"

	"zgame/internet/ssoauth/internal/util"

	"github.com/gin-gonic/gin"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Success bool   `json:"success"`
	Session string `json:"session"`
	Message string `json:"message"`
}

// Login 处理登录请求
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	// 简单的用户名密码验证（实际项目中应从数据库验证）
	if req.Username == "admin" && req.Password == "password" {
		// 生成32位小写字母的session
		session := util.GenerateSession()

		// 生成JWT token（实际项目中应存储到数据库或缓存）
		token, _ := util.GenerateJWT(req.Username)
		_ = token // 这里可以存储token

		c.JSON(http.StatusOK, LoginResponse{
			Success: true,
			Session: session,
			Message: "Login successful",
		})
	} else {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		})
	}
}