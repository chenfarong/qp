package bill

import (
	"testing"

	"github.com/aoyo/qp/internal/bill/model"
	"github.com/aoyo/qp/internal/bill/service"
	"github.com/aoyo/qp/pkg/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTokenService(t *testing.T) {
	// 连接测试数据库
	uri := "mongodb://admin:password@localhost:27017/qp_game?authSource=admin"
	dbInstance, err := db.InitDB(uri)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbInstance.Close()

	// 初始化代币服务
	tokenService := service.NewTokenService(dbInstance, "qp_game")

	// 初始化代币类型
	tokenType := model.TokenType{
		Name:        "金币",
		Symbol:      "GOLD",
		Type:        "currency",
		Description: "游戏金币",
	}

	err = tokenService.InitTokenType(tokenType)
	if err != nil {
		t.Logf("Token type already exists: %v", err)
	}

	// 生成测试用户ID
	userID := primitive.NewObjectID().Hex()

	// 测试获取用户代币余额
	userToken, err := tokenService.GetUserToken(userID, "GOLD")
	if err != nil {
		t.Fatalf("Failed to get user token: %v", err)
	}

	if userToken.Balance != 0 {
		t.Fatalf("Initial balance should be 0, got %d", userToken.Balance)
	}

	// 测试增加用户代币
	err = tokenService.AddUserToken(userID, "GOLD", 100)
	if err != nil {
		t.Fatalf("Failed to add user token: %v", err)
	}

	// 测试获取用户代币余额
	userToken, err = tokenService.GetUserToken(userID, "GOLD")
	if err != nil {
		t.Fatalf("Failed to get user token: %v", err)
	}

	if userToken.Balance != 100 {
		t.Fatalf("Balance should be 100, got %d", userToken.Balance)
	}

	// 测试减少用户代币
	err = tokenService.RemoveUserToken(userID, "GOLD", 50)
	if err != nil {
		t.Fatalf("Failed to remove user token: %v", err)
	}

	// 测试获取用户代币余额
	userToken, err = tokenService.GetUserToken(userID, "GOLD")
	if err != nil {
		t.Fatalf("Failed to get user token: %v", err)
	}

	if userToken.Balance != 50 {
		t.Fatalf("Balance should be 50, got %d", userToken.Balance)
	}

	// 测试锁定用户代币
	err = tokenService.LockUserToken(userID, "GOLD", 20)
	if err != nil {
		t.Fatalf("Failed to lock user token: %v", err)
	}

	// 测试获取用户代币余额
	userToken, err = tokenService.GetUserToken(userID, "GOLD")
	if err != nil {
		t.Fatalf("Failed to get user token: %v", err)
	}

	if userToken.Balance != 30 || userToken.Locked != 20 {
		t.Fatalf("Balance should be 30 and locked should be 20, got balance: %d, locked: %d", userToken.Balance, userToken.Locked)
	}

	// 测试解锁用户代币
	err = tokenService.UnlockUserToken(userID, "GOLD", 10)
	if err != nil {
		t.Fatalf("Failed to unlock user token: %v", err)
	}

	// 测试获取用户代币余额
	userToken, err = tokenService.GetUserToken(userID, "GOLD")
	if err != nil {
		t.Fatalf("Failed to get user token: %v", err)
	}

	if userToken.Balance != 40 || userToken.Locked != 10 {
		t.Fatalf("Balance should be 40 and locked should be 10, got balance: %d, locked: %d", userToken.Balance, userToken.Locked)
	}

	t.Log("Token service tests passed!")
}

func TestPaymentService(t *testing.T) {
	// 连接测试数据库
	uri := "mongodb://admin:password@localhost:27017/qp_game?authSource=admin"
	dbInstance, err := db.InitDB(uri)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbInstance.Close()

	// 初始化代币服务
	tokenService := service.NewTokenService(dbInstance, "qp_game")
	// 初始化支付服务
	paymentService := service.NewPaymentService(dbInstance, "qp_game", tokenService)

	// 生成测试用户ID
	userID := primitive.NewObjectID().Hex()

	// 测试创建支付订单
	payment := model.Payment{
		UserID:      userID,
		Amount:      1000,
		Currency:    "CNY",
		TokenType:   "GOLD",
		TokenAmount: 10000,
		PaymentMethod: "wechat",
	}

	createdPayment, err := paymentService.CreatePayment(payment)
	if err != nil {
		t.Fatalf("Failed to create payment: %v", err)
	}

	if createdPayment.OrderID == "" {
		t.Fatalf("Order ID should not be empty")
	}

	if createdPayment.Status != "pending" {
		t.Fatalf("Status should be pending, got %s", createdPayment.Status)
	}

	// 测试获取支付订单
	getPayment, err := paymentService.GetPayment(createdPayment.OrderID)
	if err != nil {
		t.Fatalf("Failed to get payment: %v", err)
	}

	if getPayment.OrderID != createdPayment.OrderID {
		t.Fatalf("Order ID should match")
	}

	// 测试生成支付URL
	paymentURL, err := paymentService.GeneratePaymentURL(createdPayment.OrderID, "wechat")
	if err != nil {
		t.Fatalf("Failed to generate payment URL: %v", err)
	}

	if paymentURL == "" {
		t.Fatalf("Payment URL should not be empty")
	}

	// 测试查询支付状态
	status, err := paymentService.QueryPaymentStatus(createdPayment.OrderID)
	if err != nil {
		t.Fatalf("Failed to query payment status: %v", err)
	}

	if status != "pending" {
		t.Fatalf("Status should be pending, got %s", status)
	}

	// 测试处理支付回调
	err = paymentService.HandlePaymentCallback(createdPayment.OrderID, "trans_123", "success")
	if err != nil {
		t.Fatalf("Failed to handle payment callback: %v", err)
	}

	// 测试查询支付状态
	status, err = paymentService.QueryPaymentStatus(createdPayment.OrderID)
	if err != nil {
		t.Fatalf("Failed to query payment status: %v", err)
	}

	if status != "success" {
		t.Fatalf("Status should be success, got %s", status)
	}

	// 测试用户代币余额
	userToken, err := tokenService.GetUserToken(userID, "GOLD")
	if err != nil {
		t.Fatalf("Failed to get user token: %v", err)
	}

	if userToken.Balance != 10000 {
		t.Fatalf("Balance should be 10000, got %d", userToken.Balance)
	}

	t.Log("Payment service tests passed!")
}
