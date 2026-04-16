package base

import (
	"context"
	pb "zagame/pb/golang/gamelogic"
)

// Handler 基础处理器
type Handler struct {
	Service *Service
}

// Login 登录请求处理
func (h *Handler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return h.Service.Login(ctx, req)
}

// GetRoleInfo 获取角色信息请求处理
func (h *Handler) GetRoleInfo(ctx context.Context, req *pb.GetRoleInfoRequest) (*pb.GetRoleInfoResponse, error) {
	return h.Service.GetRoleInfo(ctx, req)
}

// GetGameMoney 获取游戏货币请求处理
func (h *Handler) GetGameMoney(ctx context.Context, req *pb.GetGameMoneyRequest) (*pb.GetGameMoneyResponse, error) {
	return h.Service.GetGameMoney(ctx, req)
}

// NewHandler 创建基础处理器
func NewHandler() *Handler {
	return &Handler{
		Service: NewService(),
	}
}
