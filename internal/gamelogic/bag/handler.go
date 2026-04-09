package bag

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// InventoryHandler 背包处理器
type InventoryHandler struct {
	inventoryService *InventoryService
}

// NewInventoryHandler 创建背包处理器实例
func NewInventoryHandler(inventoryService *InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

// GetInventory 获取角色背包（与角色同一文档）
func (h *InventoryHandler) GetInventory(c *gin.Context) {
	characterID := c.Query("character_id")
	if characterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing character_id"})
		return
	}

	resp, err := h.inventoryService.GetInventory(characterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AddItem 添加物品到背包
func (h *InventoryHandler) AddItem(c *gin.Context) {
	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.inventoryService.AddItem(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item added successfully"})
}

// UseItem 使用物品
func (h *InventoryHandler) UseItem(c *gin.Context) {
	var req UseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.inventoryService.UseItem(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item used successfully"})
}

// RemoveItem 从背包中删除物品
func (h *InventoryHandler) RemoveItem(c *gin.Context) {
	var req RemoveItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.inventoryService.RemoveItem(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item removed successfully"})
}

// EquipItem 装备或卸下物品
func (h *InventoryHandler) EquipItem(c *gin.Context) {
	var req EquipItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.inventoryService.EquipItem(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item equipped successfully"})
}

// RegisterRoutes 注册路由
func (h *InventoryHandler) RegisterRoutes(router *gin.RouterGroup) {
	inventoryGroup := router.Group("/inventory")
	{
		inventoryGroup.GET("", h.GetInventory)
		inventoryGroup.POST("/items", h.AddItem)
		inventoryGroup.POST("/items/use", h.UseItem)
		inventoryGroup.POST("/items/remove", h.RemoveItem)
		inventoryGroup.POST("/items/equip", h.EquipItem)
	}
}
