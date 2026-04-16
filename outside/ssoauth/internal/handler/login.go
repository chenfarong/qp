package handler

import (
	"log"
	"net/http"
	"sync"

	"zgame/internet/ssoauth/internal/util"

	"github.com/gin-gonic/gin"
)

// 内存存储用户信息
type User struct {
	Username string
	Password string
}

var (
	users      = make(map[string]*User)
	usersMutex sync.RWMutex
)

// CheckAndCreateDefaultAdmin 检查并创建默认的za_admin账号
func CheckAndCreateDefaultAdmin() {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	// 检查za_admin账号是否存在
	if _, exists := users["za_admin"]; !exists {
		// 创建za_admin账号，密码和账号一样
		users["za_admin"] = &User{
			Username: "za_admin",
			Password: "za_admin",
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
	usersMutex.RLock()
	user, exists := users[req.Username]
	usersMutex.RUnlock()

	if !exists || user.Password != req.Password {
		log.Printf("登录失败: 用户名或密码错误: %s\n", req.Username)
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
	usersMutex.Lock()
	if _, exists := users[req.Username]; exists {
		usersMutex.Unlock()
		log.Printf("注册失败: 用户名已存在: %s\n", req.Username)
		c.JSON(http.StatusConflict, RegisterResponse{
			Success: false,
			Message: "Username already exists",
		})
		return
	}

	// 存储用户信息到内存
	users[req.Username] = &User{
		Username: req.Username,
		Password: req.Password,
	}
	usersMutex.Unlock()

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
