package handler

import (
	"net/http"

	"github.com/aoyo/qp/internal/chat/service"
	"github.com/gin-gonic/gin"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	chatService *service.ChatService
}

// NewChatHandler 创建聊天处理器实例
func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// SendMessage 发送消息
func (h *ChatHandler) SendMessage(c *gin.Context) {
	var req service.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.chatService.SendMessage(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMessages 获取消息历史
func (h *ChatHandler) GetMessages(c *gin.Context) {
	var req service.GetMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	messages, err := h.chatService.GetMessages(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// GetConversations 获取会话列表
func (h *ChatHandler) GetConversations(c *gin.Context) {
	var req service.GetConversationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversations, err := h.chatService.GetConversations(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

// UpdateMessageStatus 更新消息状态
func (h *ChatHandler) UpdateMessageStatus(c *gin.Context) {
	var req service.UpdateMessageStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.chatService.UpdateMessageStatus(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}

// RegisterRoutes 注册路由
func (h *ChatHandler) RegisterRoutes(router *gin.Engine) {
	chatGroup := router.Group("/api/chat")
	{
		chatGroup.POST("/messages", h.SendMessage)
		chatGroup.POST("/messages/history", h.GetMessages)
		chatGroup.POST("/conversations", h.GetConversations)
		chatGroup.POST("/messages/status", h.UpdateMessageStatus)
	}
}
