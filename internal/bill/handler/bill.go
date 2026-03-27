package handler

import (
	"net/http"

	"github.com/aoyo/qp/internal/bill/model"
	"github.com/aoyo/qp/internal/bill/service"
	"github.com/gin-gonic/gin"
)

// BillHandler 账单处理器
type BillHandler struct {
	tokenService    *service.TokenService
	paymentService  *service.PaymentService
}

// NewBillHandler 创建账单处理器实例
func NewBillHandler(tokenService *service.TokenService, paymentService *service.PaymentService) *BillHandler {
	return &BillHandler{
		tokenService:    tokenService,
		paymentService:  paymentService,
	}
}

// RegisterRoutes 注册路由
func (h *BillHandler) RegisterRoutes(router *gin.Engine) {
	bill := router.Group("/bill")
	{
		// 代币相关接口
		token := bill.Group("/token")
		{
			token.GET("/balance", h.GetUserToken)
			token.POST("/add", h.AddUserToken)
			token.POST("/remove", h.RemoveUserToken)
			token.POST("/lock", h.LockUserToken)
			token.POST("/unlock", h.UnlockUserToken)
		}
		
		// 支付相关接口
		payment := bill.Group("/payment")
		{
			payment.POST("/create", h.CreatePayment)
			payment.GET("/get", h.GetPayment)
			payment.POST("/callback", h.HandlePaymentCallback)
			payment.GET("/status", h.QueryPaymentStatus)
			payment.GET("/url", h.GeneratePaymentURL)
		}
	}
}

// GetUserToken 获取用户代币余额
func (h *BillHandler) GetUserToken(c *gin.Context) {
	userID := c.Query("user_id")
	tokenType := c.Query("token_type")
	
	if userID == "" || tokenType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id or token_type"})
		return
	}
	
	userToken, err := h.tokenService.GetUserToken(userID, tokenType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, userToken)
}

// AddUserToken 增加用户代币
func (h *BillHandler) AddUserToken(c *gin.Context) {
	var req struct {
		UserID    string `json:"user_id" binding:"required"`
		TokenType string `json:"token_type" binding:"required"`
		Amount    int64  `json:"amount" binding:"required,gt=0"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := h.tokenService.AddUserToken(req.UserID, req.TokenType, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "token added successfully"})
}

// RemoveUserToken 减少用户代币
func (h *BillHandler) RemoveUserToken(c *gin.Context) {
	var req struct {
		UserID    string `json:"user_id" binding:"required"`
		TokenType string `json:"token_type" binding:"required"`
		Amount    int64  `json:"amount" binding:"required,gt=0"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := h.tokenService.RemoveUserToken(req.UserID, req.TokenType, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "token removed successfully"})
}

// LockUserToken 锁定用户代币
func (h *BillHandler) LockUserToken(c *gin.Context) {
	var req struct {
		UserID    string `json:"user_id" binding:"required"`
		TokenType string `json:"token_type" binding:"required"`
		Amount    int64  `json:"amount" binding:"required,gt=0"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := h.tokenService.LockUserToken(req.UserID, req.TokenType, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "token locked successfully"})
}

// UnlockUserToken 解锁用户代币
func (h *BillHandler) UnlockUserToken(c *gin.Context) {
	var req struct {
		UserID    string `json:"user_id" binding:"required"`
		TokenType string `json:"token_type" binding:"required"`
		Amount    int64  `json:"amount" binding:"required,gt=0"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := h.tokenService.UnlockUserToken(req.UserID, req.TokenType, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "token unlocked successfully"})
}

// CreatePayment 创建支付订单
func (h *BillHandler) CreatePayment(c *gin.Context) {
	var payment model.Payment
	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	createdPayment, err := h.paymentService.CreatePayment(payment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, createdPayment)
}

// GetPayment 获取支付订单
func (h *BillHandler) GetPayment(c *gin.Context) {
	orderID := c.Query("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing order_id"})
		return
	}
	
	payment, err := h.paymentService.GetPayment(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, payment)
}

// HandlePaymentCallback 处理支付回调
func (h *BillHandler) HandlePaymentCallback(c *gin.Context) {
	var req struct {
		OrderID       string `json:"order_id" binding:"required"`
		TransactionID string `json:"transaction_id" binding:"required"`
		Status        string `json:"status" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := h.paymentService.HandlePaymentCallback(req.OrderID, req.TransactionID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "callback handled successfully"})
}

// QueryPaymentStatus 查询支付状态
func (h *BillHandler) QueryPaymentStatus(c *gin.Context) {
	orderID := c.Query("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing order_id"})
		return
	}
	
	status, err := h.paymentService.QueryPaymentStatus(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status": status})
}

// GeneratePaymentURL 生成支付URL
func (h *BillHandler) GeneratePaymentURL(c *gin.Context) {
	orderID := c.Query("order_id")
	paymentMethod := c.Query("payment_method")
	
	if orderID == "" || paymentMethod == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing order_id or payment_method"})
		return
	}
	
	url, err := h.paymentService.GeneratePaymentURL(orderID, paymentMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"url": url})
}
