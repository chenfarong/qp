package battle

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BattleHandler 战斗处理器
type BattleHandler struct {
	battleService *BattleService
}

// NewBattleHandler 创建战斗处理器实例
func NewBattleHandler(battleService *BattleService) *BattleHandler {
	return &BattleHandler{
		battleService: battleService,
	}
}

// Battle 战斗
func (h *BattleHandler) Battle(c *gin.Context) {
	var req BattleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.battleService.Battle(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RegisterRoutes 注册路由
func (h *BattleHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/battle", h.Battle)
}
