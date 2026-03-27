package ssoauth

import (
	"testing"

	"github.com/aoyo/qp/internal/ssoauth/service"
	"github.com/aoyo/qp/pkg/db"
)

func TestAuthService(t *testing.T) {
	// 连接测试数据库
	uri := "mongodb://admin:password@localhost:27017/qp_game?authSource=admin"
	dbInstance, err := db.InitDB(uri)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbInstance.Close()

	// 初始化认证服务
	authService := service.NewAuthService(dbInstance, "test_secret", 24, "qp_game")

	// 测试注册功能
	registerReq := service.RegisterRequest{
		Username: "newuser",
		Password: "password123",
		Email:    "newuser@example.com",
		Nickname: "New User",
	}

	registerResp, err := authService.Register(registerReq)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	if registerResp.Token == "" {
		t.Fatalf("Register response should contain token")
	}

	// 测试登录功能
	loginReq := service.LoginRequest{
		Username: registerReq.Username,
		Password: registerReq.Password,
	}

	loginResp, err := authService.Login(loginReq)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	if loginResp.Token == "" {
		t.Fatalf("Login response should contain token")
	}

	// 测试验证令牌
	claims, err := authService.ValidateToken(loginResp.Token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID == "" {
		t.Fatalf("Token should contain user ID")
	}

	t.Log("Auth service tests passed!")
}
