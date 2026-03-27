package handler

import (
	"net/http"

	"github.com/aoyo/qp/internal/gamelogic/service"
	"github.com/gin-gonic/gin"
)

// GameHandler 游戏处理器
type GameHandler struct {
	gameService *service.GameService
}

// NewGameHandler 创建游戏处理器实例
func NewGameHandler(gameService *service.GameService) *GameHandler {
	return &GameHandler{
		gameService: gameService,
	}
}

// CreateCharacter 创建角色
func (h *GameHandler) CreateCharacter(c *gin.Context) {
	var req service.CreateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.gameService.CreateCharacter(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetCharacters 获取用户的所有角色
func (h *GameHandler) GetCharacters(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
		return
	}

	characters, err := h.gameService.GetCharactersByUserID(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"characters": characters})
}

// GetCharacter 获取角色详情
func (h *GameHandler) GetCharacter(c *gin.Context) {
	characterIDStr := c.Param("id")
	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character_id"})
		return
	}

	character, err := h.gameService.GetCharacterByID(characterIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	c.JSON(http.StatusOK, character)
}

// UpdateCharacterStatus 更新角色状态
func (h *GameHandler) UpdateCharacterStatus(c *gin.Context) {
	characterIDStr := c.Param("id")
	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character_id"})
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.gameService.UpdateCharacterStatus(characterIDStr, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}

// Battle 战斗
func (h *GameHandler) Battle(c *gin.Context) {
	var req service.BattleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.gameService.Battle(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RegisterRoutes 注册路由
func (h *GameHandler) RegisterRoutes(router *gin.Engine) {
	gameGroup := router.Group("/api/game")
	{
		gameGroup.POST("/characters", h.CreateCharacter)
		gameGroup.GET("/characters", h.GetCharacters)
		gameGroup.GET("/characters/:id", h.GetCharacter)
		gameGroup.PUT("/characters/:id/status", h.UpdateCharacterStatus)
		gameGroup.POST("/battle", h.Battle)
	}
}
