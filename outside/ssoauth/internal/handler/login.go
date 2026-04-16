package handler

import (
	"database/sql"
	"log"
	"net/http"

	"zgame/database"
	"zgame/internet/ssoauth/internal/util"

	"github.com/gin-gonic/gin"
)

// CheckAndCreateDefaultAdmin 检查并创建默认的za_admin账号
func CheckAndCreateDefaultAdmin() {
	// 检查za_admin账号是否存在
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", "za_admin").Scan(&count)
	if err != nil {
		log.Printf("检查za_admin账号失败: %v\n", err)
		return
	}

	if count == 0 {
		// 创建za_admin账号，密码和账号一样
		_, err := database.DB.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", "za_admin", "za_admin")
		if err != nil {
			log.Printf("创建za_admin账号失败: %v\n", err)
			return
		}
		log.Println("Created default za_admin account")
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
		log.Printf("登录请求参数错误: %v\n", err)
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	log.Printf("收到登录请求: username=%s\n", req.Username)

	// 验证用户名和密码
	var password string
	err := database.DB.QueryRow("SELECT password FROM users WHERE username = $1", req.Username).Scan(&password)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("登录失败: 用户名不存在: %s\n", req.Username)
			c.JSON(http.StatusUnauthorized, LoginResponse{
				Success: false,
				Message: "Invalid username or password",
			})
			return
		}
		log.Printf("登录失败: 数据库错误: %v\n", err)
		c.JSON(http.StatusInternalServerError, LoginResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}

	if password != req.Password {
		log.Printf("登录失败: 密码错误: %s\n", req.Username)
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		})
		return
	}

	// 生成32位小写字母和数字的session
	session := util.GenerateSession()

	// 生成JWT token
	token, _ := util.GenerateJWT(req.Username)

	log.Printf("登录成功: username=%s\n", req.Username)

	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Session: session,
		Token:   token,
		Message: "Login successful",
	})
}

// Register 处理注册请求
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("注册请求参数错误: %v\n", err)
		c.JSON(http.StatusBadRequest, RegisterResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	log.Printf("收到注册请求: username=%s\n", req.Username)

	// 检查用户名是否已存在
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", req.Username).Scan(&count)
	if err != nil {
		log.Printf("注册失败: 数据库错误: %v\n", err)
		c.JSON(http.StatusInternalServerError, RegisterResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}

	if count > 0 {
		log.Printf("注册失败: 用户名已存在: %s\n", req.Username)
		c.JSON(http.StatusConflict, RegisterResponse{
			Success: false,
			Message: "Username already exists",
		})
		return
	}

	// 存储用户信息到数据库
	_, err = database.DB.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", req.Username, req.Password)
	if err != nil {
		log.Printf("注册失败: 数据库错误: %v\n", err)
		c.JSON(http.StatusInternalServerError, RegisterResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}

	log.Printf("注册成功: username=%s\n", req.Username)

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
		log.Println("获取个人信息失败: 用户未认证")
		c.JSON(http.StatusUnauthorized, ProfileResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	log.Printf("获取个人信息成功: username=%s\n", username.(string))

	c.JSON(http.StatusOK, ProfileResponse{
		Success:  true,
		Username: username.(string),
		Message:  "Profile retrieved successfully",
	})
}
