package handler

import (
	"log"
	"net/http"

	"github.com/aoyo/qp/internal/ssoauth/service"
	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register 注册处理
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Register(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建会话
	sessionReq := service.SessionRequest{
		UserID:    resp.UserInfo.ID.Hex(),
		Token:     resp.Token,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
	_, err = h.authService.CreateSession(sessionReq)
	if err != nil {
		// 会话创建失败不影响注册
		log.Println("Failed to create session:", err)
	}

	c.JSON(http.StatusOK, resp)
}

// Login 登录处理
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建会话
	sessionReq := service.SessionRequest{
		UserID:    resp.UserInfo.ID.Hex(),
		Token:     resp.Token,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
	_, err = h.authService.CreateSession(sessionReq)
	if err != nil {
		// 会话创建失败不影响登录
		log.Println("Failed to create session:", err)
	}

	c.JSON(http.StatusOK, resp)
}

// Validate 验证令牌
func (h *AuthHandler) Validate(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	// 移除Bearer前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// 验证JWT令牌
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// 检查会话是否存在且有效
	_, err = h.authService.GetSession(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired or not found"})
		return
	}

	user, err := h.authService.GetUserByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": claims.UserID,
		"user":    user,
	})
}

// Refresh 刷新令牌
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建新会话
	sessionReq := service.SessionRequest{
		UserID:    resp.UserInfo.ID.Hex(),
		Token:     resp.Token,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
	_, err = h.authService.CreateSession(sessionReq)
	if err != nil {
		// 会话创建失败不影响令牌刷新
		log.Println("Failed to create session:", err)
	}

	c.JSON(http.StatusOK, resp)
}

// Logout 注销处理
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
		return
	}

	// 移除Bearer前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// 删除会话
	err := h.authService.DeleteSession(token)
	if err != nil {
		// 会话删除失败不影响注销
		log.Println("Failed to delete session:", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

// RegisterRoutes 注册路由
func (h *AuthHandler) RegisterRoutes(router *gin.Engine) {
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.GET("/validate", h.Validate)
		authGroup.POST("/refresh", h.Refresh)
		authGroup.POST("/logout", h.Logout)
	}
}
