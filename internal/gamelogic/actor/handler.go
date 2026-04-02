package actor

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CharacterHandler 角色处理器
type CharacterHandler struct {
	characterService *CharacterService
}

// NewCharacterHandler 创建角色处理器实例
func NewCharacterHandler(characterService *CharacterService) *CharacterHandler {
	return &CharacterHandler{
		characterService: characterService,
	}
}

// CreateCharacter 创建角色
func (h *CharacterHandler) CreateCharacter(c *gin.Context) {
	var req CreateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.characterService.CreateCharacter(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetCharacters 获取用户的所有角色
func (h *CharacterHandler) GetCharacters(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
		return
	}

	characters, err := h.characterService.GetCharactersByUserID(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"characters": characters})
}

// GetCharacter 获取角色详情
func (h *CharacterHandler) GetCharacter(c *gin.Context) {
	characterIDStr := c.Param("id")
	if characterIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character_id"})
		return
	}

	character, err := h.characterService.GetCharacterByID(characterIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	c.JSON(http.StatusOK, character)
}

// UpdateCharacterStatus 更新角色状态
func (h *CharacterHandler) UpdateCharacterStatus(c *gin.Context) {
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

	if err := h.characterService.UpdateCharacterStatus(characterIDStr, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}

// UseCharacter 使用角色
func (h *CharacterHandler) UseCharacter(c *gin.Context) {
	var req UseCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.characterService.UseCharacter(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RegisterRoutes 注册路由
func (h *CharacterHandler) RegisterRoutes(router *gin.RouterGroup) {
	characterGroup := router.Group("/characters")
	{
		characterGroup.POST("", h.CreateCharacter)
		characterGroup.GET("", h.GetCharacters)
		characterGroup.GET("/:id", h.GetCharacter)
		characterGroup.PUT("/:id/status", h.UpdateCharacterStatus)
		characterGroup.POST("/use", h.UseCharacter)
	}
}
