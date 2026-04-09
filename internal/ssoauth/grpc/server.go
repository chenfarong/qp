package grpc

import (
	"context"

	"github.com/aoyo/qp/internal/ssoauth/service"
	"github.com/aoyo/qp/pkg/proto/auth"
)

// AuthServer 认证服务gRPC服务器
type AuthServer struct {
	auth.UnimplementedAuthServiceServer
	authService *service.AuthService
}

// NewAuthServer 创建认证服务gRPC服务器实例
func NewAuthServer(authService *service.AuthService) *AuthServer {
	return &AuthServer{
		authService: authService,
	}
}

// Register 用户注册
func (s *AuthServer) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	// 转换请求参数
	registerReq := service.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Nickname: req.Nickname,
	}

	// 调用服务
	resp, err := s.authService.Register(registerReq)
	if err != nil {
		return &auth.RegisterResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	user := &auth.User{
		Id:       1, // 临时值
		Username: resp.UserInfo.Username,
		Email:    resp.UserInfo.Email,
		Nickname: resp.UserInfo.Nickname,
		Avatar:   resp.UserInfo.Avatar,
		Status:   int32(resp.UserInfo.Status),
	}

	return &auth.RegisterResponse{
		Token:    resp.Token,
		ExpireAt: resp.ExpireAt,
		UserInfo: user,
	}, nil
}

// Login 用户登录
func (s *AuthServer) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	// 转换请求参数
	loginReq := service.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	// 调用服务
	resp, err := s.authService.Login(loginReq)
	if err != nil {
		return &auth.LoginResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	user := &auth.User{
		Id:       1, // 临时值
		Username: resp.UserInfo.Username,
		Email:    resp.UserInfo.Email,
		Nickname: resp.UserInfo.Nickname,
		Avatar:   resp.UserInfo.Avatar,
		Status:   int32(resp.UserInfo.Status),
	}

	return &auth.LoginResponse{
		Token:    resp.Token,
		ExpireAt: resp.ExpireAt,
		UserInfo: user,
	}, nil
}

// Validate 验证令牌
func (s *AuthServer) Validate(ctx context.Context, req *auth.ValidateRequest) (*auth.ValidateResponse, error) {
	// 调用服务
	claims, err := s.authService.ValidateToken(req.Token)
	if err != nil {
		return &auth.ValidateResponse{
			Error: err.Error(),
		}, nil
	}

	// 根据用户ID获取用户信息
	user, err := s.authService.GetUserByID(claims.UserID)
	if err != nil {
		return &auth.ValidateResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	authUser := &auth.User{
		Id:       1, // 临时值
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Status:   int32(user.Status),
	}

	return &auth.ValidateResponse{
		UserId: 1, // 临时值
		User:   authUser,
	}, nil
}
