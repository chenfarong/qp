package handler

import (
	"fmt"
	"net/http"

	"zgame/internet/ssoauth/internal/util"

	"github.com/gin-gonic/gin"
)

// 存储用户信息（实际项目中应该使用数据库）
var users = map[string]string{
	"admin": "password", // 初始用户
}

// 初始化函数，检查并创建默认的za_admin账号
func init() {
	// 检查za_admin账号是否存在
	if _, exists := users["za_admin"]; !exists {
		// 创建za_admin账号，密码和账号一样
		users["za_admin"] = "za_admin"
		fmt.Println("Created default za_admin account")
	}
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Success bool   `json:"success"`
	Session string `json:"session"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterResponse 注册响应结构
type RegisterResponse struct {
	Success bool   `json:"success"`
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

	// 验证用户名和密码
	if password, exists := users[req.Username]; exists && password == req.Password {
		// 生成32位小写字母和数字的session
		session := util.GenerateSession()

		// 生成JWT token
		token, _ := util.GenerateJWT(req.Username)

		c.JSON(http.StatusOK, LoginResponse{
			Success: true,
			Session: session,
			Token:   token,
			Message: "Login successful",
		})
	} else {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		})
	}
}

// Register 处理注册请求
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RegisterResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	// 检查用户名是否已存在
	if _, exists := users[req.Username]; exists {
		c.JSON(http.StatusConflict, RegisterResponse{
			Success: false,
			Message: "Username already exists",
		})
		return
	}

	// 存储用户信息
	users[req.Username] = req.Password

	c.JSON(http.StatusOK, RegisterResponse{
		Success: true,
		Message: "Registration successful",
	})
}

// ProfileResponse 个人信息响应结构
type ProfileResponse struct {
	Success  bool   `json:"success"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// Profile 处理个人信息请求
func Profile(c *gin.Context) {
	// 从上下文中获取用户名
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, ProfileResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	c.JSON(http.StatusOK, ProfileResponse{
		Success:  true,
		Username: username.(string),
		Message:  "Profile retrieved successfully",
	})
}
